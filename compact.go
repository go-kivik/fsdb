package fs

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-kivik/fsdb/decoder"
	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/fsdb/internal"
)

type docEntry struct {
	Meta           *internal.DocMeta
	AttachmentsDir string
	Revs           docIndex
}

func newEntry() *docEntry {
	return &docEntry{
		Meta: &internal.DocMeta{},
		Revs: make(docIndex),
	}
}

type docIndex map[string]*docEntry

// get gets the docEntry associated with docID, creating it if necessary.
func (i docIndex) get(docID string) *docEntry {
	e, ok := i[docID]
	if ok {
		return e
	}
	e = newEntry()
	i[docID] = e
	return e
}

func (i docIndex) readRoot(ctx context.Context, fs filesystem.Filesystem, path string) error {
	return i.readIndex(ctx, fs, path, true)
}

func (i docIndex) readRevs(ctx context.Context, fs filesystem.Filesystem, path string) error {
	return i.readIndex(ctx, fs, path, false)
}

func (i docIndex) readIndex(ctx context.Context, fs filesystem.Filesystem, path string, root bool) error {
	dir, err := fs.Open(path)
	if err != nil {
		return kerr(err)
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return kerr(err)
	}

	for _, info := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !info.IsDir() {
			id, ext, ok := explodeFilename(info.Name())
			if !ok {
				// ignore unrecognized files
				continue
			}
			var docID string
			var rev internal.Rev
			if root {
				docID, err = filename2id(id)
				if err != nil {
					// ignore unrecognized files
					continue
				}
			} else {
				if err := rev.UnmarshalText([]byte(id)); err != nil {
					// ignore unrecognized files
					continue
				}
			}
			entry := i.get(id)
			entry.Meta, err = decoder.ReadDocMeta(fs, filepath.Join(path, id), ext)
			if err != nil {
				return err
			}
			if docID != "" {
				entry.Meta.ID = docID
			}
			if rev.Seq != 0 {
				entry.Meta.Rev = rev
			}
			entry.Meta.Path = filepath.Join(path, info.Name())
			continue
		}
		if root && info.Name()[0] == '.' {
			id := strings.TrimPrefix(info.Name(), ".")
			entry := i.get(id)
			if err := entry.Revs.readRevs(ctx, fs, filepath.Join(path, info.Name())); err != nil {
				return err
			}
			continue
		}
		entry := i.get(info.Name())
		entry.AttachmentsDir = filepath.Join(path, info.Name())
	}
	return i.joinWinningRevs()
}

// joinWinningRevs looks for any winning revs that have attachments left in the
// revs dir, and joins them, so such attachments are not considered abandoned.
func (i docIndex) joinWinningRevs() error {
	for _, entry := range i {
		if entry.Meta.Path == "" {
			continue
		}
		if len(entry.Revs) == 0 {
			continue
		}
		doc, err := readDoc(entry.Meta.Path)
		if err != nil {
			return err
		}
		rev := doc.Rev.String()
		if revEntry, ok := entry.Revs[rev]; ok {
			if revEntry.Meta.Path == "" {
				revEntry.Meta.Path = entry.Meta.Path
			}
		}
	}
	return nil
}

func (i docIndex) removeAbandonedAttachments(fs filesystem.Filesystem) error {
	for _, entry := range i {
		if entry.AttachmentsDir != "" {
			if entry.Meta.Path == "" {
				if err := os.RemoveAll(entry.AttachmentsDir); err != nil {
					return kerr(err)
				}
				continue
			}
			doc, err := readDoc(entry.Meta.Path)
			if err != nil {
				return err
			}
			attDir, err := fs.Open(entry.AttachmentsDir)
			if err != nil {
				return kerr(err)
			}
			atts, err := attDir.Readdir(-1)
			if err != nil {
				return kerr(err)
			}
			for _, att := range atts {
				if _, ok := doc.Attachments[att.Name()]; !ok {
					if err := os.Remove(filepath.Join(entry.AttachmentsDir, att.Name())); err != nil {
						return kerr(err)
					}
				}
			}
		}
		if err := entry.Revs.removeAbandonedAttachments(fs); err != nil {
			return err
		}
	}
	return nil
}

func explodeFilename(filename string) (basename, ext string, ok bool) {
	dotExt := filepath.Ext(filename)
	basename = strings.TrimSuffix(filename, dotExt)
	ext = strings.TrimPrefix(dotExt, ".")
	for _, e := range decoder.Extensions() {
		if e == ext {
			return basename, ext, true
		}
	}
	return "", "", false
}

func (d *db) Compact(ctx context.Context) error {
	return d.compact(ctx, filesystem.Default())
}

func (d *db) compact(ctx context.Context, fs filesystem.Filesystem) error {
	docs := docIndex{}
	if err := docs.readRoot(ctx, fs, d.path()); err != nil {
		return err
	}

	return docs.removeAbandonedAttachments(fs)
}
