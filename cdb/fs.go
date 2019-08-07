package cdb

import (
	"path/filepath"

	"github.com/go-kivik/fsdb/cdb/decode"
	"github.com/go-kivik/fsdb/filesystem"
)

// FS provides filesystem access to a
type FS struct {
	fs   filesystem.Filesystem
	root string
}

// New initializes a new FS instance, anchored at dbroot. If fs is omitted or
// nil, the default is used.
func New(dbroot string, fs ...filesystem.Filesystem) *FS {
	var vfs filesystem.Filesystem
	if len(fs) > 0 {
		vfs = fs[0]
	}
	if vfs == nil {
		vfs = filesystem.Default()
	}
	return &FS{
		fs:   vfs,
		root: dbroot,
	}
}

// Open opens the requested document.
func (fs *FS) Open(docID string) (*Document, error) {
	base := escapeID(docID)
	f, ext, err := decode.OpenAny(fs.fs, filepath.Join(fs.root, base))
	if err != nil {
		return nil, kerr(missing(err))
	}
	rev := new(Revision)
	if err := decode.Decode(f, ext, rev); err != nil {
		return nil, err
	}
	doc := &Document{
		ID:        docID,
		Revisions: []*Revision{rev},
	}
	return doc, nil
}
