package fs

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-kivik/kivik"
	"github.com/go-kivik/kivik/driver"
)

// TODO:
// - atts_since
// - conflicts
// - deleted_conflicts
// - latest
// - local_seq
// - meta
// - open_revs
func (d *db) Get(ctx context.Context, docID string, opts map[string]interface{}) (*driver.Document, error) {
	if docID == "" {
		return nil, &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "no docid specified"}
	}
	doc, err := d.cdb.OpenDocID(docID, opts)
	if err != nil {
		return nil, err
	}
	doc.Options = opts
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(doc); err != nil {
		return nil, err
	}
	attsIter, err := doc.Revisions[0].AttachmentsIterator()
	if err != nil {
		return nil, err
	}
	return &driver.Document{
		Rev:           doc.Revisions[0].Rev.String(),
		Body:          ioutil.NopCloser(buf),
		ContentLength: int64(buf.Len()),
		Attachments:   attsIter,
	}, nil
}
