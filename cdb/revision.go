package cdb

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-kivik/fsdb/filesystem"
)

// RevMeta is the metadata stored in reach revision.
type RevMeta struct {
	Rev         RevID                  `json:"_rev" yaml:"_rev"`
	Deleted     *bool                  `json:"_deleted,omitempty" yaml:"_deleted,omitempty"`
	Attachments map[string]*Attachment `json:"_attachments,omitempty" yaml:"_attachments,omitempty"`
	RevHistory  *RevHistory            `json:"_revisions,omitempty" yaml:"_revisions,omitempty"`

	path string
	fs   filesystem.Filesystem
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
	for key := range reservedKeys {
		delete(r.Data, key)
	}
	return nil
}

// UnmarshalYAML satisfies the yaml.Unmarshaler interface.
func (r *Revision) UnmarshalYAML(u func(interface{}) error) error {
	if err := u(&r.RevMeta); err != nil {
		return err
	}
	if err := u(&r.Data); err != nil {
		return err
	}
	for key := range reservedKeys {
		delete(r.Data, key)
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
	result := make([]byte, 0, len(dataJSON)+len(metaJSON)-1)
	result = append(result, metaJSON[:len(metaJSON)-1]...)
	result = append(result, ',')
	result = append(result, dataJSON[1:]...)
	return result, nil
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
