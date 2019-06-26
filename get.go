package fs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"

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
// - rev
// - revs
// - revs_info
func (d *db) Get(_ context.Context, docID string, opts map[string]interface{}) (*driver.Document, error) {
	if docID == "" {
		return nil, &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "no docid specified"}
	}
	filename := id2filename(docID)
	base := base(filename)
	if rev, ok := opts["rev"].(string); ok {
		filename = "." + base + "/" + rev + path.Ext(filename)
	}
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
	ndoc := new(normalDoc)
	if err := json.NewDecoder(f).Decode(&ndoc); err != nil {
		return nil, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}
	doc := &driver.Document{
		ContentLength: stat.Size(),
		Body:          f,
		Rev:           ndoc.Rev,
	}
	if ok, _ := opts["attachments"].(bool); ok {
		_ = f.Close()
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
		r, w := io.Pipe()
		go func() {
			err := json.NewEncoder(w).Encode(ndoc)
			w.CloseWithError(err) // nolint: errcheck
		}()
		doc.Body = r
		doc.Attachments = atts
	}
	return doc, nil
}
