package cdb

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

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

func (fs *FS) readMainRev(base string) (*Revision, error) {
	f, ext, err := decode.OpenAny(fs.fs, base)
	if err != nil {
		return nil, kerr(missing(err))
	}
	defer f.Close() // nolint: errcheck
	rev := new(Revision)
	if err := decode.Decode(f, ext, rev); err != nil {
		return nil, err
	}
	rev.path = base + "." + ext
	rev.fs = fs.fs
	if rev.Rev.IsZero() {
		rev.Rev = RevID{Seq: 1}
	}
	return rev, nil
}

func (fs *FS) readSubRev(path string) (*Revision, error) {
	ext := filepath.Ext(path)
	var revid RevID
	basename := filepath.Base(strings.TrimSuffix(path, ext))
	if err := revid.UnmarshalText([]byte(basename)); err != nil {
		// skip unrecognized files
		return nil, errUnrecognizedFile
	}

	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, kerr(missing(err))
	}
	defer f.Close() // nolint: errcheck
	rev := new(Revision)
	if err := decode.Decode(f, ext, rev); err != nil {
		return nil, err
	}
	rev.path = path
	rev.fs = fs.fs
	rev.Rev = revid
	return rev, nil
}

func (fs *FS) openRevs(docID string) ([]*Revision, error) {
	revs := make(Revisions, 0, 1)
	base := escapeID(docID)
	rev, err := fs.readMainRev(filepath.Join(fs.root, base))
	if err != nil && err != errNotFound {
		return nil, err
	}
	if err == nil {
		revs = append(revs, rev)
	}
	dirpath := filepath.Join(fs.root, "."+base)
	dir, err := fs.fs.Open(dirpath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		files, err := dir.Readdir(-1)
		if err != nil {
			return nil, err
		}
		for _, info := range files {
			if info.IsDir() {
				continue
			}
			rev, err := fs.readSubRev(filepath.Join(dirpath, info.Name()))
			switch {
			case err == errUnrecognizedFile:
				continue
			case err != nil:
				return nil, err
			}
			revs = append(revs, rev)
		}
	}
	if len(revs) == 0 {
		return nil, errNotFound
	}
	sort.Sort(revs)
	return revs, nil
}

// Open opens the requested document.
func (fs *FS) Open(docID string) (*Document, error) {
	revs, err := fs.openRevs(docID)
	if err != nil {
		return nil, err
	}
	doc := &Document{
		ID:        docID,
		Revisions: revs,
	}
	for _, rev := range doc.Revisions {
		for filename, att := range rev.Attachments {
			path, file, err := rev.openAttachment(filename)
			if err != nil {
				return nil, err
			}
			_ = file.Close()
			att.path = path
			att.fs = fs.fs
		}
	}
	return doc, nil
}
