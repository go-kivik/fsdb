package fs

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-kivik/fsdb/decoder"
	"github.com/go-kivik/kivik"
)

func filename2id(filename string) (string, error) {
	return url.PathUnescape(filename)
}

func (d *db) currentRev(docID, ext string) (string, error) {
	f, err := os.Open(d.path(docID))
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return "", err
	}
	r, err := decoder.Rev(f, ext)
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

var reservedPrefixes = []string{"_local/", "_design/"}

func validateID(id string) error {
	if id[0] != '_' {
		return nil
	}
	for _, prefix := range reservedPrefixes {
		if strings.HasPrefix(id, prefix) && len(id) > len(prefix) {
			return nil
		}
	}
	return &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "only reserved document ids may start with underscore"}
}

/*
File naming strategy:
Current rev lives under:    {db}/{docid}.{ext}
Historical revs live under: {db}/.{docid}/{rev}
Attachments:                {db}/{docid}/{filename}
*/
func (d *db) Put(ctx context.Context, docID string, i interface{}, opts map[string]interface{}) (string, error) {
	if err := validateID(docID); err != nil {
		return "", err
	}
	rev, err := d.cdb.NewRevision(i)
	if err != nil {
		return "", err
	}
	doc, err := d.cdb.OpenDocID(docID, opts)
	switch {
	case kivik.StatusCode(err) == http.StatusNotFound:
		// Crate new doc
		doc = d.cdb.NewDocument(docID)
	case err != nil:
		return "", err
	}
	return doc.AddRevision(ctx, rev, opts)
}
