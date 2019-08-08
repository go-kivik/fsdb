package cdb

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-kivik/fsdb/filesystem"
)

// RevMeta is the metadata stored in reach revision.
type RevMeta struct {
	Rev         RevID                  `json:"_rev" yaml:"_rev"`
	Deleted     *bool                  `json:"_deleted,omitempty" yaml:"_deleted,omitempty"`
	Attachments map[string]*Attachment `json:"_attachments,omitempty" yaml:"_attachments,omitempty"`
	RevHistory  *RevHistory            `json:"_revisions,omitempty" yaml:"_revisions,omitempty"`

	// isMain should be set to true when unmarshaling the main Rev, to enable
	// auto-population of the _rev key, if necessary
	isMain bool
	path   string
	fs     filesystem.Filesystem
}

// Revision is a specific instance of a document.
type Revision struct {
	RevMeta

	// Data is the normal payload
	Data map[string]interface{} `json:"-" yaml:"-"`
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
	metaJSON, err := json.Marshal(r.RevMeta)
	if err != nil {
		return nil, err
	}
	if len(r.Data) == 0 {
		return metaJSON, nil
	}
	dataJSON, err := json.Marshal(r.Data)
	if err != nil {
		return nil, err
	}
	return joinJSON(metaJSON, dataJSON), nil
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
		f, err := r.fs.Open(filepath.Join(path, rev, filename))
		if !os.IsNotExist(err) {
			return filepath.Join(path, rev, filename), f, err
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
