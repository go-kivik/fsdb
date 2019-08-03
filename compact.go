package fs

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-kivik/fsdb/decoder"
)

type docIndex struct {
	doc           string
	attachmentDir string
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

	docs := map[string]*docIndex{}
	for _, i := range files {
		if !i.IsDir() {
			docID, _, ok := explodeFilename(i.Name())
			if !ok {
				// ignore unrecognized files
				continue
			}
			di, ok := docs[docID]
			if !ok {
				di = &docIndex{}
				docs[docID] = di
			}
			di.doc = d.path(i.Name())
			continue
		}
		di, ok := docs[i.Name()]
		if !ok {
			di = &docIndex{}
			docs[i.Name()] = di
		}
		di.attachmentDir = d.path(i.Name())
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
