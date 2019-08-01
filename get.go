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

const (
	// Maybe make this confiurable at some point?
	revsLimit = 1000
)

type revsInfo struct {
	Rev    string `json:"rev"`
	Status string `json:"status"`
}

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
	ndoc, err := d.get(ctx, docID, opts)
	if err != nil {
		return nil, err
	}
	doc := &driver.Document{
		Rev: ndoc.Rev.String(),
	}
	if _, ok := opts["rev"]; ok {
		delete(ndoc.Data, "_revisions")
	} else {
		if ok, _ := opts["revs_info"].(bool); ok {
			ndoc.Data["_revs_info"] = ndoc.revsInfo()
		}
		if ok, _ := opts["revs"].(bool); ok {
			if _, ok := ndoc.Data["_revisions"]; !ok {
				histSize := ndoc.Rev.Seq
				if histSize > revsLimit {
					histSize = revsLimit
				}
				var ids []string
				if ndoc.Rev.Sum == "" {
					ids = make([]string, int(histSize))
				} else {
					ids = []string{ndoc.Rev.Sum}
				}
				ndoc.Data["_revisions"] = map[string]interface{}{
					"start": ndoc.Rev.Seq,
					"ids":   ids,
				}
			}
		}
	}
	for _, att := range ndoc.Attachments {
		if att.RevPos == 0 {
			att.RevPos = ndoc.Rev.Seq
		}
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
				RevPos:      att.RevPos,
			}
		}
		doc.Attachments = atts
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
