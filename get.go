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
	"github.com/go-kivik/fsdb/internal"
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
	doc, err := d.cdb.Open(docID)
	if err != nil {
		return nil, err
	}
	// ndoc, err := d.get(ctx, docID, opts)
	// if err != nil {
	// 	return nil, err
	// }
	// for _, att := range ndoc.Attachments {
	// 	if att.RevPos == 0 {
	// 		att.RevPos = ndoc.Rev.Seq
	// 	}
	// }
	// if ok, _ := opts["attachments"].(bool); ok {
	// 	atts := make(internal.Attachments)
	// 	for filename, att := range ndoc.Attachments {
	// 		f, err := d.openAttachment(ctx, docID, ndoc.Revisions, filename)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		info, err := f.Stat()
	// 		if err != nil {
	// 			return nil, kerr(err)
	// 		}
	// 		att.Stub = false
	// 		att.Follows = true
	// 		att.Content = f
	// 		atts[filename] = &internal.Attachment{
	// 			Content:     f,
	// 			Size:        info.Size(),
	// 			ContentType: att.ContentType,
	// 			Digest:      att.ContentType,
	// 			RevPos:      att.RevPos,
	// 		}
	// 	}
	// 	doc.Attachments = atts
	// }
	// if ndoc.ID != docID {
	// 	ndoc.ID = docID
	// }
	// if _, ok := opts["revs"]; !ok {
	// 	ndoc.Revisions = nil
	// }
	// if _, ok := opts["rev"]; ok {
	// 	ndoc.Revisions = nil
	// } else {
	// 	if ok, _ := opts["revs_info"].(bool); ok {
	// 		ndoc.RevsInfo = ndoc.GetRevsInfo()
	// 	}
	// }
	doc.Options = opts
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(doc); err != nil {
		return nil, err
	}
	return &driver.Document{
		Rev:           doc.Revisions[0].Rev.String(),
		Body:          ioutil.NopCloser(buf),
		ContentLength: int64(buf.Len()),
	}, nil
}

func (d *db) openAttachment(ctx context.Context, docID string, revs *internal.Revisions, filename string) (filesystem.File, error) {
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
