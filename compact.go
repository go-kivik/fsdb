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

func (i docIndex) readRoot(path string) error {
	return i.readIndex(path, true)
}

func (i docIndex) readRevs(path string) error {
	return i.readIndex(path, false)
}

func (i docIndex) readIndex(path string, root bool) error {
	dir, err := os.Open(path)
	if err != nil {
		return kerr(err)
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return kerr(err)
	}

	for _, info := range files {
		if !info.IsDir() {
			id, _, ok := explodeFilename(info.Name())
			if !ok {
				// ignore unrecognized files
				continue
			}
			entry := i.get(id)
			entry.doc = filepath.Join(path, info.Name())
			continue
		}
		if root && info.Name()[0] == '.' {
			id := strings.TrimPrefix(info.Name(), ".")
			entry := i.get(id)
			if err := entry.revs.readRevs(filepath.Join(path, info.Name())); err != nil {
				return err
			}
			continue
		}
		entry := i.get(info.Name())
		entry.attachmentDir = filepath.Join(path, info.Name())
	}
	return i.joinWinningRevs()
}

// joinWinningRevs looks for any winning revs that have attachments left in the
// revs dir, and joins them, so such attachments are not considered abandoned.
func (i docIndex) joinWinningRevs() error {
	for _, entry := range i {
		if entry.doc == "" {
			continue
		}
		if len(entry.revs) == 0 {
			continue
		}
		doc, err := readDoc(entry.doc)
		if err != nil {
			return err
		}
		rev := doc.Rev.String()
		if revEntry, ok := entry.revs[rev]; ok {
			if revEntry.doc == "" {
				revEntry.doc = entry.doc
			}
		}
	}
	return nil
}

func (i docIndex) removeAbandonedAttachments() error {
	for _, entry := range i {
		if entry.doc == "" && entry.attachmentDir != "" {
			if err := os.RemoveAll(entry.attachmentDir); err != nil {
				return kerr(err)
			}
		}
		if err := entry.revs.removeAbandonedAttachments(); err != nil {
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
	docs := docIndex{}
	if err := docs.readRoot(d.path()); err != nil {
		return err
	}

	return docs.removeAbandonedAttachments()
}
