package cdb

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/kivik"
	"github.com/icza/dyno"
)

// RevMeta is the metadata stored in reach revision.
type RevMeta struct {
	Rev         RevID                  `json:"_rev" yaml:"_rev"`
	Deleted     *bool                  `json:"_deleted,omitempty" yaml:"_deleted,omitempty"`
	Attachments map[string]*Attachment `json:"_attachments,omitempty" yaml:"_attachments,omitempty"`
	RevHistory  *RevHistory            `json:"_revisions,omitempty" yaml:"_revisions,omitempty"`

	// isMain should be set to true when unmarshaling the main Rev, to enable
	// auto-population of the _rev key, if necessary
	isMain bool                  // nolint: structcheck
	path   string                // nolint: structcheck
	fs     filesystem.Filesystem // nolint: structcheck
}

// Revision is a specific instance of a document.
type Revision struct {
	RevMeta

	// Data is the normal payload
	Data map[string]interface{} `json:"-" yaml:"-"`

	options kivik.Options
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (r *Revision) UnmarshalJSON(p []byte) error {
	if err := json.Unmarshal(p, &r.RevMeta); err != nil {
		return err
	}
	if err := json.Unmarshal(p, &r.Data); err != nil {
		return err
	}
	return r.finalizeUnmarshal()
}

// UnmarshalYAML satisfies the yaml.Unmarshaler interface.
func (r *Revision) UnmarshalYAML(u func(interface{}) error) error {
	if err := u(&r.RevMeta); err != nil {
		return err
	}
	if err := u(&r.Data); err != nil {
		return err
	}
	r.Data = dyno.ConvertMapI2MapS(r.Data).(map[string]interface{})
	return r.finalizeUnmarshal()
}

func (r *Revision) finalizeUnmarshal() error {
	for key := range reservedKeys {
		delete(r.Data, key)
	}
	if r.isMain && r.Rev.IsZero() {
		r.Rev = RevID{Seq: 1}
	}
	if !r.isMain && r.path != "" {
		revstr := filepath.Base(strings.TrimSuffix(r.path, filepath.Ext(r.path)))
		if err := r.Rev.UnmarshalText([]byte(revstr)); err != nil {
			return errUnrecognizedFile
		}
	}
	if r.RevHistory == nil {
		var ids []string
		if r.Rev.Sum == "" {
			histSize := r.Rev.Seq
			if histSize > revsLimit {
				histSize = revsLimit
			}
			ids = make([]string, int(histSize))
		} else {
			ids = []string{r.Rev.Sum}
		}
		r.RevHistory = &RevHistory{
			Start: r.Rev.Seq,
			IDs:   ids,
		}
	}
	return nil
}

// MarshalJSON satisfies the json.Marshaler interface
func (r *Revision) MarshalJSON() ([]byte, error) {
	var meta interface{} = r.RevMeta
	revs, _ := r.options["revs"].(bool)
	if _, ok := r.options["rev"]; ok {
		revs = false
	}
	if !revs {
		meta = struct {
			RevMeta
			// This suppresses RevHistory from being included in the default output
			RevHistory *RevHistory `json:"_revisions,omitempty"` // nolint: govet
		}{
			RevMeta: r.RevMeta,
		}
	}
	stub, follows := r.stubFollows()
	for _, att := range r.Attachments {
		att.outputStub = stub
		att.Follows = follows
	}
	parts := make([]json.RawMessage, 0, 2)
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	parts = append(parts, metaJSON)
	if len(r.Data) > 0 {
		dataJSON, err := json.Marshal(r.Data)
		if err != nil {
			return nil, err
		}
		parts = append(parts, dataJSON)
	}
	return joinJSON(parts...), nil
}

func (r *Revision) stubFollows() (bool, bool) {
	attachments, _ := r.options["attachments"].(bool)
	if !attachments {
		return true, false
	}
	accept, _ := r.options["header:accept"].(string)
	return false, accept != "application/json"
}

func (r *Revision) openAttachment(filename string) (filesystem.File, error) {
	path := strings.TrimSuffix(r.path, filepath.Ext(r.path))
	f, err := r.fs.Open(filepath.Join(path, filename))
	if !os.IsNotExist(err) {
		return f, err
	}
	basename := filepath.Base(path)
	path = strings.TrimSuffix(path, basename)
	if basename != r.Rev.String() {
		// We're working with the main rev
		path += "." + basename
	}
	for _, rev := range r.RevHistory.Ancestors() {
		fullpath := filepath.Join(path, rev, filename)
		f, err := r.fs.Open(fullpath)
		if !os.IsNotExist(err) {
			return f, err
		}
	}
	return nil, errNotFound
}

// Revisions is a sortable list of document revisions.
type Revisions []*Revision

var _ sort.Interface = Revisions{}

// Len returns the number of elements in r.
func (r Revisions) Len() int {
	return len(r)
}

func (r Revisions) Less(i, j int) bool {
	return r[i].Rev.Seq > r[j].Rev.Seq ||
		(r[i].Rev.Seq == r[j].Rev.Seq && r[i].Rev.Sum > r[j].Rev.Sum)
}

func (r Revisions) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// Delete deletes the revision.
func (r *Revision) Delete(ctx context.Context) error {
	if err := os.Remove(r.path); err != nil {
		return err
	}
	attpath := strings.TrimSuffix(r.path, filepath.Ext(r.path))
	return os.RemoveAll(attpath)
}

// NewRevision creates a new revision from i, according to opts.
func (fs *FS) NewRevision(i interface{}) (*Revision, error) {
	data, err := json.Marshal(i)
	if err != nil {
		return nil, &kivik.Error{HTTPStatus: http.StatusBadRequest, Err: err}
	}
	rev := new(Revision)
	rev.fs = fs.fs
	if err := json.Unmarshal(data, &rev); err != nil {
		return nil, &kivik.Error{HTTPStatus: http.StatusBadRequest, Err: err}
	}
	for _, att := range rev.Attachments {
		if att.RevPos == nil {
			revpos := rev.Rev.Seq
			att.RevPos = &revpos
		}
	}
	return rev, nil
}

func (r *Revision) persist(ctx context.Context, path string) error {
	if err := r.fs.Mkdir(filepath.Dir(path), 0777); err != nil && !os.IsExist(err) {
		return err
	}
	var dirMade bool
	for attname, att := range r.Attachments {
		if att.Stub || att.path != "" {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if !dirMade {
			if err := r.fs.Mkdir(path, 0777); err != nil && !os.IsExist(err) {
				return err
			}
			dirMade = true
		}
		target := filepath.Join(path, escapeID(attname))
		if err := atomicWriteFile(r.fs, target, bytes.NewReader(att.Content)); err != nil {
			return err
		}
		att.path = target
	}
	f := atomicFileWriter(r.fs, path+".json")
	defer f.Close() // nolint: errcheck
	r.options = kivik.Options{"revs": true}
	if err := json.NewEncoder(f).Encode(r); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	r.path = path + ".json"
	return nil
}

// hash passes deterministic JSON content of the revision through md5 to
// generate a hash to be used in the revision ID.
func (r *Revision) hash() (string, error) {
	r.options = nil
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	h := md5.New()
	_, _ = h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
