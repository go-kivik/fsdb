package fs

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-kivik/fsdb/decoder"
	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/fsdb/internal"
	"github.com/go-kivik/kivik"
)

func normalizeDoc(i interface{}) (*internal.Document, error) {
	data, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	doc := &internal.Document{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, &kivik.Error{HTTPStatus: http.StatusBadRequest, Err: err}
	}
	return doc, nil
}

func readDoc(path string) (*internal.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, kerr(err)
	}
	ext := filepath.Ext(path)[1:]
	i, err := decoder.Decode(f, ext)
	if err != nil {
		return nil, err
	}
	doc, err := normalizeDoc(i)
	if err != nil {
		return nil, err
	}
	doc.Path = path
	return doc, nil
}

func (d *db) openDoc(docID, rev string) (filesystem.File, string, error) {
	base := id2basename(docID)
	for _, ext := range decoder.Extensions() {
		filename := base + "." + ext
		if rev != "" {
			currev, err := d.currentRev(filename, ext)
			if err != nil && !os.IsNotExist(err) {
				return nil, "", err
			}
			if currev != rev {
				revFilename := "." + base + "/" + rev + "." + ext
				f, err := d.fs.Open(d.path(revFilename))
				if !os.IsNotExist(err) {
					return f, ext, err
				}
				continue
			}
		}
		f, err := d.fs.Open(d.path(filename))
		if !os.IsNotExist(err) {
			return f, ext, kerr(err)
		}
	}
	if rev == "" {
		return breakRevTie(d.path("." + base))
	}
	return nil, "", errNotFound
}

func breakRevTie(path string) (*os.File, string, error) {
	dir, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = errNotFound
		}
		return nil, "", kerr(err)
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, "", kerr(err)
	}
	revs := make(internal.DocMetas, 0, len(files))
	for _, info := range files {
		if info.IsDir() {
			continue
		}
		ext := filepath.Ext(info.Name())
		base := strings.TrimSuffix(info.Name(), ext)
		rev := internal.RevID{}
		if err := rev.UnmarshalText([]byte(base)); err != nil {
			// Ignore unrecognized files
			continue
		}
		revs = append(revs, &internal.DocMeta{
			Rev:  rev,
			Path: filepath.Join(path, info.Name()),
		})
	}

	if len(revs) == 0 {
		return nil, "", errNotFound
	}

	sort.Sort(revs)

	winner := revs[0]

	f, err := os.Open(winner.Path)
	return f, winner.Ext(), kerr(err)
}

func (d *db) readDoc(docID, rev string) (*internal.Document, error) {
	f, ext, err := d.openDoc(docID, rev)
	if err != nil {
		return nil, err
	}
	defer f.Close() // nolint: errcheck
	i, err := decoder.Decode(f, ext)
	if err != nil {
		return nil, err
	}
	doc, err := normalizeDoc(i)
	if err != nil {
		return nil, err
	}
	doc.Path = f.Name()
	return doc, nil
}

func (d *db) get(_ context.Context, docID string, opts map[string]interface{}) (*internal.Document, error) {
	rev, _ := opts["rev"].(string)
	ndoc, err := d.readDoc(docID, rev)
	if err != nil {
		return nil, kerr(err)
	}
	if ndoc.Rev.IsZero() {
		ndoc.Rev.Increment()
	}
	ndoc.Revisions = ndoc.GetRevisions()
	return ndoc, nil
}
