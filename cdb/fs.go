package cdb

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-kivik/fsdb/v4/cdb/decode"
	"github.com/go-kivik/fsdb/v4/filesystem"
	"github.com/go-kivik/kivik/v4"
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
	rev.isMain = true
	rev.path = base + "." + ext
	rev.fs = fs.fs
	if err := decode.Decode(f, ext, rev); err != nil {
		return nil, err
	}
	if err := rev.restoreAttachments(); err != nil {
		return nil, err
	}
	return rev, nil
}

func (fs *FS) readSubRev(path string) (*Revision, error) {
	ext := filepath.Ext(path)

	f, err := fs.fs.Open(path)
	if err != nil {
		return nil, kerr(missing(err))
	}
	defer f.Close() // nolint: errcheck
	rev := new(Revision)
	rev.path = path
	rev.fs = fs.fs
	if err := decode.Decode(f, ext, rev); err != nil {
		return nil, err
	}
	if err := rev.restoreAttachments(); err != nil {
		return nil, err
	}
	return rev, nil
}

func (r *Revision) restoreAttachments() error {
	for attname, att := range r.Attachments {
		if att.RevPos == nil {
			revpos := r.Rev.Seq
			att.RevPos = &revpos
		}
		if att.Size == 0 || att.Digest == "" {
			f, err := r.openAttachment(attname)
			if err != nil {
				return &kivik.Error{HTTPStatus: http.StatusInternalServerError, Err: err}
			}
			att.Size, att.Digest = digest(f)
			_ = f.Close()
		}
	}
	return nil
}

func (fs *FS) openRevs(docID, revid string) ([]*Revision, error) {
	revs := make(Revisions, 0, 1)
	base := EscapeID(docID)
	rev, err := fs.readMainRev(filepath.Join(fs.root, base))
	if err != nil && err != errNotFound {
		return nil, err
	}
	if err == nil {
		if revid == "" || rev.Rev.String() == revid {
			revs = append(revs, rev)
		}
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
			if revid != "" {
				base := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
				if base != revid {
					continue
				}
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

// OpenDocID opens the requested document by ID (without file extension).
func (fs *FS) OpenDocID(docID string, opts kivik.Options) (*Document, error) {
	rev, _ := opts["rev"].(string)
	revs, err := fs.openRevs(docID, rev)
	if err != nil {
		return nil, err
	}
	doc := &Document{
		ID:        docID,
		Revisions: revs,
		cdb:       fs,
	}
	for _, rev := range doc.Revisions {
		for filename, att := range rev.Attachments {
			file, err := rev.openAttachment(filename)
			if err != nil {
				return nil, err
			}
			_ = file.Close()
			att.path = file.Name()
			att.fs = fs.fs
		}
	}
	return doc, nil
}
