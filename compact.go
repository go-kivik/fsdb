package fs

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/go-kivik/fsdb/cdb"
	"github.com/go-kivik/fsdb/decoder"
	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/fsdb/internal"
)

type docEntry struct {
	internal.DocMeta
	Revs docIndex
}

func newEntry() *docEntry {
	return &docEntry{
		Revs: make(docIndex),
	}
}

type docIndex map[string]*cdb.Document

func (i docIndex) readIndex(ctx context.Context, fs filesystem.Filesystem, path string) error {
	dir, err := fs.Open(path)
	if err != nil {
		return kerr(err)
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return kerr(err)
	}

	c := cdb.New(path, fs)

	var docID string
	for _, info := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		switch {
		case !info.IsDir():
			id, _, ok := explodeFilename(info.Name())
			if !ok {
				// ignore unrecognized files
				continue
			}
			docID = id
		case info.IsDir() && info.Name()[0] == '.':
			docID = strings.TrimPrefix(info.Name(), ".")
		default:
			continue
		}
		if _, ok := i[docID]; ok {
			// We've already read this one
			continue
		}
		doc, err := c.OpenDocID(docID, nil)
		if err != nil {
			return err
		}
		i[docID] = doc
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
	if err := docs.readIndex(ctx, fs, d.path()); err != nil {
		return err
	}
	for _, doc := range docs {
		if err := doc.Compact(ctx); err != nil {
			return err
		}
	}
	return nil
}
