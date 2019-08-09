package cdb

import (
	"context"
	"encoding/json"
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
	for _, att := range r.Attachments {
		if att.RevPos == 0 {
			att.RevPos = r.Rev.Seq
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
		att.Stub = stub
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

func (r *Revision) openAttachment(filename string) (string, filesystem.File, error) {
	path := strings.TrimSuffix(r.path, filepath.Ext(r.path))
	f, err := r.fs.Open(filepath.Join(path, filename))
	if !os.IsNotExist(err) {
		return filepath.Join(path, filename), f, err
	}
	basename := filepath.Base(path)
	path = strings.TrimSuffix(path, basename)
	if basename != r.Rev.String() {
		// We're working with the main rev
		path += "." + basename
	}
	for _, rev := range r.RevHistory.ancestors() {
		fullpath := filepath.Join(path, rev, filename)
		f, err := r.fs.Open(fullpath)
		if !os.IsNotExist(err) {
			return fullpath, f, err
		}
	}
	return "", nil, errNotFound
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
func NewRevision(i interface{}) (*Revision, error) {
	data, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	rev := new(Revision)
	err = json.Unmarshal(data, &rev)
	return rev, err
}
