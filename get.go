package fs

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

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
// - revs
// - revs_info
func (d *db) Get(_ context.Context, docID string, opts map[string]interface{}) (*driver.Document, error) {
	if docID == "" {
		return nil, &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "no docid specified"}
	}
	rev, _ := opts["rev"].(string)
	ndoc, err := d.readDoc(docID, rev)
	if err != nil {
		if os.IsPermission(err) {
			return nil, &kivik.Error{HTTPStatus: http.StatusForbidden, Err: err}
		}
		return nil, err
	}
	doc := &driver.Document{
		Rev: ndoc.Rev.String(),
	}
	if ok, _ := opts["attachments"].(bool); ok {
		base := strings.TrimPrefix(base(ndoc.Path), d.path())
		atts := make(attachments)
		for filename, att := range ndoc.Attachments {
			f, err := os.Open(d.path(base, filename))
			if err != nil {
				return nil, err
			}
			att.Stub = false
			att.Follows = true
			att.Content = f
			atts[filename] = &attachment{
				Content:     f,
				Size:        att.Size,
				ContentType: att.ContentType,
				Digest:      att.ContentType,
			}
		}
		doc.Attachments = atts
	}
	if ndoc.Rev.IsZero() {
		ndoc.Rev.Increment()
	}
	if ndoc.ID != docID {
		ndoc.ID = docID
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(ndoc); err != nil {
		return nil, err
	}
	doc.Body = ioutil.NopCloser(buf)
	doc.Rev = ndoc.Rev.String()
	doc.ContentLength = int64(buf.Len())
	return doc, nil
}
