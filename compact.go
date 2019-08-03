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
	revs          docIndex
}

func newEntry() *docEntry {
	return &docEntry{
		revs: make(docIndex),
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
			revdir, err := os.Open(d.path(i.Name()))
			if err != nil {
				return err
			}
			revFiles, err := revdir.Readdir(-1)
			if err != nil {
				return err
			}
			for _, ri := range revFiles {
				if !ri.IsDir() {
					rev, _, ok := explodeFilename(ri.Name())
					if !ok {
						// ignore unrecognized files
						continue
					}
					re := e.revs.get(rev)
					re.doc = d.path(i.Name(), ri.Name())
					continue
				}
				re := e.revs.get(i.Name())
				re.attachmentDir = d.path(i.Name(), ri.Name())
			}
			continue
		}
		e := docs.get(i.Name())
		e.attachmentDir = d.path(i.Name())
	}

	for _, e := range docs {
		if e.doc != "" {
			continue
		}
		if e.attachmentDir != "" {
			if err := os.RemoveAll(e.attachmentDir); err != nil {
				return err
			}
		}
		for _, re := range e.revs {
			if re.doc != "" {
				continue
			}
			if re.attachmentDir != "" {
				if err := os.RemoveAll(re.attachmentDir); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
