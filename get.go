package fs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/go-kivik/kivik"
	"github.com/go-kivik/kivik/driver"
)

// This should be optimized to only read the first few bytes of the file.
func readRev(f io.ReadSeeker) (rev string, err error) {
	doc := struct {
		Rev string `json:"_rev"`
	}{}
	defer func() {
		_, e := f.Seek(0, 0)
		if err == nil {
			err = e
		}
	}()
	err = json.NewDecoder(f).Decode(&doc)
	return doc.Rev, err
}

func (d *db) Get(_ context.Context, docID string, opts map[string]interface{}) (*driver.Document, error) {
	if docID == "" {
		return nil, &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "no docid specified"}
	}
	filename := id2filename(docID)
	f, err := os.Open(d.path(filename))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &kivik.Error{HTTPStatus: http.StatusNotFound, Err: err}
		}
		if os.IsPermission(err) {
			return nil, &kivik.Error{HTTPStatus: http.StatusForbidden, Err: err}
		}
		return nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	rev, err := readRev(f)
	if err != nil {
		return nil, err
	}
	return &driver.Document{
		ContentLength: stat.Size(),
		Body:          f,
		Rev:           rev,
	}, nil
}
