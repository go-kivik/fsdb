package fs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-kivik/fsdb/filesystem"
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
	for _, att := range ndoc.Attachments {
		if att.RevPos == 0 {
			att.RevPos = ndoc.Rev.Seq
		}
	}
	if ok, _ := opts["attachments"].(bool); ok {
		atts := make(attachments)
		for filename, att := range ndoc.Attachments {
			f, err := d.openAttachment(ctx, docID, ndoc.Revisions, filename)
			if err != nil {
				return nil, err
			}
			info, err := f.Stat()
			if err != nil {
				return nil, kerr(err)
			}
			att.Stub = false
			att.Follows = true
			att.Content = f
			atts[filename] = &attachment{
				Content:     f,
				Size:        info.Size(),
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
	if _, ok := opts["revs"]; !ok {
		ndoc.Revisions = nil
	}
	if _, ok := opts["rev"]; ok {
		ndoc.Revisions = nil
	} else {
		if ok, _ := opts["revs_info"].(bool); ok {
			ndoc.RevsInfo = ndoc.revsInfo()
		}
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

func (d *db) openAttachment(ctx context.Context, docID string, revs *revisions, filename string) (filesystem.File, error) {
	f, err := d.fs.Open(d.path(docID, filename))
	if err == nil {
		return f, nil
	}
	if !os.IsNotExist(err) {
		return f, kerr(err)
	}
	for i, revid := range revs.IDs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		rev := fmt.Sprintf("%d-%s", revs.Start-int64(i), revid)
		f, err := os.Open(d.path("."+docID, rev, filename))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return f, kerr(err)
		}
		return f, nil
	}
	return nil, errNotFound
}
