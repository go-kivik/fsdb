package fs

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-kivik/fsdb/decoder"
)

type docEntry struct {
	doc           string
	attachmentDir string
	revs          []*docEntry
}

type docIndex map[string]*docEntry

// get gets the docEntry associated with docID, creating it if necessary.
func (i docIndex) get(docID string) *docEntry {
	e, ok := i[docID]
	if ok {
		return e
	}
	e = &docEntry{}
	i[docID] = e
	return e
}

func explodeFilename(filename string) (basename, ext string, ok bool) {
	dotExt := filepath.Ext(filename)
	basename = strings.TrimSuffix(filename, dotExt)
	ext = strings.TrimPrefix(filename, ".")
	for _, e := range decoder.Extensions() {
		if e == ext {
			return basename, ext, true
		}
	}
	return "", "", false
}

func (d *db) Compact(ctx context.Context) error {
	dir, err := os.Open(d.path())
	if err != nil {
		return kerr(err)
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	docs := docIndex{}
	for _, i := range files {
		if !i.IsDir() {
			docID, _, ok := explodeFilename(i.Name())
			if !ok {
				// ignore unrecognized files
				continue
			}
			e := docs.get(docID)
			e.doc = d.path(i.Name())
			continue
		}
		if i.Name()[0] == '.' {
			docID := strings.TrimPrefix(i.Name(), ".")
			e := docs.get(docID)
			e.revs = append(e.revs, &docEntry{
				doc: d.path(i.Name()),
			})
			continue
		}
		e := docs.get(i.Name())
		e.attachmentDir = d.path(i.Name())
	}

	for _, idx := range docs {
		if idx.doc != "" {
			continue
		}
		if idx.attachmentDir != "" {
			if err := os.RemoveAll(idx.attachmentDir); err != nil {
				return err
			}
		}
	}

	return nil
}
