package fs

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
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
	f, err := d.openDoc(docID, rev)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &kivik.Error{HTTPStatus: http.StatusNotFound, Err: err}
		}
		if os.IsPermission(err) {
			return nil, &kivik.Error{HTTPStatus: http.StatusForbidden, Err: err}
		}
		return nil, err
	}
	defer f.Close()
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
		Rev:           ndoc.Rev.String(),
	}
	if ok, _ := opts["attachments"].(bool); ok {
		base := strings.TrimPrefix(base(f.Name()), d.path())
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

func (d *db) openDoc(docID, rev string) (*os.File, error) {
	base := id2basename(docID)
	filename := base + ".json"
	if rev != "" {
		currev, err := d.currentRev(filename)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if currev != rev {
			revFilename := "." + base + "/" + rev + path.Ext(filename)
			return os.Open(d.path(revFilename))
		}
	}
	return os.Open(d.path(filename))
}
