package cdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-kivik/kivik"
	"golang.org/x/xerrors"
)

// Document is a CouchDB document.
type Document struct {
	ID        string    `json:"_id" yaml:"_id"`
	Revisions Revisions `json:"-" yaml:"-"`
	// RevsInfo is only used during JSON marshaling, and should never be
	// consulted as authoritative.
	RevsInfo []RevInfo `json:"_revs_info,omitempty" yaml:"-"`

	Options kivik.Options `json:"-" yaml:"-"`

	// path is the path to the database
	path string
	fs   *FS
}

// NewDocument creates a new document.
func (fs *FS) NewDocument(path, docID string) *Document {
	return &Document{
		ID:   docID,
		path: path,
		fs:   fs,
	}
}

// MarshalJSON satisfies the json.Marshaler interface.
func (d *Document) MarshalJSON() ([]byte, error) {
	d.revsInfo()
	rev := d.Revisions[0]
	rev.options = d.Options
	revJSON, err := json.Marshal(rev)
	if err != nil {
		return nil, err
	}
	docJSON, _ := json.Marshal(*d)
	return joinJSON(docJSON, revJSON), nil
}

// revsInfo populates the RevsInfo field, if appropriate according to options.
func (d *Document) revsInfo() {
	d.RevsInfo = nil
	if ok, _ := d.Options["revs_info"].(bool); !ok {
		return
	}
	if _, ok := d.Options["rev"]; ok {
		return
	}
	d.RevsInfo = make([]RevInfo, len(d.Revisions))
	for i, rev := range d.Revisions {
		d.RevsInfo[i] = RevInfo{
			Rev:    rev.Rev.String(),
			Status: "available",
		}
	}
}

// RevInfo is revisions information as presented in the _revs_info key.
type RevInfo struct {
	Rev    string `json:"rev"`
	Status string `json:"status"`
}

// Compact cleans up any non-leaf revs, and attempts to consolidate attachments.
func (d *Document) Compact(ctx context.Context) error {
	revTree := make(map[string]*Revision, 1)
	// An index of ancestor -> leaf revision
	index := map[string][]string{}
	keep := make([]*Revision, 0, 1)
	for _, rev := range d.Revisions {
		revID := rev.Rev.String()
		if leafIDs, ok := index[revID]; ok {
			for _, leafID := range leafIDs {
				if err := copyAttachments(revTree[leafID], rev); err != nil {
					return err
				}
			}
			if err := rev.Delete(ctx); err != nil {
				return err
			}
			continue
		}
		keep = append(keep, rev)
		for _, ancestor := range rev.RevHistory.ancestors()[1:] {
			index[ancestor] = append(index[ancestor], revID)
		}
		revTree[revID] = rev
	}
	d.Revisions = keep
	return nil
}

func copyAttachments(leaf, old *Revision) error {
	leafpath := strings.TrimSuffix(leaf.path, filepath.Ext(leaf.path)) + "/"
	basepath := strings.TrimSuffix(old.path, filepath.Ext(old.path)) + "/"
	for filename, att := range old.Attachments {
		if _, ok := leaf.Attachments[filename]; !ok {
			continue
		}
		if strings.HasPrefix(att.path, basepath) {
			name := filepath.Base(att.path)
			if err := os.MkdirAll(leafpath, 0777); err != nil {
				return err
			}
			if err := os.Link(att.path, filepath.Join(leafpath, name)); err != nil {
				lerr := new(os.LinkError)
				if xerrors.As(err, &lerr) {
					if strings.HasSuffix(lerr.Error(), ": file exists") {
						if err := os.Remove(att.path); err != nil {
							return err
						}
						continue
					}
				}
				return err
			}
		}
	}
	return nil
}

// AddRevision adds rev to the existing document, according to options. The
// return value is the new revision ID.
func (d *Document) AddRevision(rev *Revision, optoins kivik.Options) (string, error) {
	d.Revisions = append(d.Revisions, rev)
	sort.Sort(d.Revisions)
	err := d.persist()
	return rev.Rev.String(), err
}

/*
Persist strategy:
- For every rev that doesn't exist on disk, create it in {db}/.{docid}/{rev}
- If winning rev does not exist in {db}/{docid}:
	- Move old winning rev to {db}/.{docid}/{rev}
	- Move new winning rev to {db}/{docid}
*/

// Document persists the contained revs to disk.
func (d *Document) persist() error {
	if d == nil || len(d.Revisions) == 0 {
		return &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "document has no revisions"}
	}
	for _, rev := range d.Revisions {
		fmt.Printf("have a rev: %s\n", rev.Rev)
		if rev.path != "" {
			continue
		}
		if err := rev.persist(filepath.Join(d.path, rev.Rev.String())); err != nil {
			return err
		}
	}

	return nil
}
