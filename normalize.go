package fs

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/go-kivik/kivik"
)

type attachments map[string]*attachment

type attachment struct {
	ContentType string   `json:"content_type"`
	Stub        bool     `json:"stub,omitempty"`
	Content     *os.File `json:"data,omitempty"`
	Size        int64    `json:"length"`
	Digest      string   `json:"digest"`
}

func (a *attachment) cleanup() error {
	if a == nil || a.Content == nil {
		return nil
	}
	if err := a.Content.Close(); err != nil {
		return err
	}
	return os.Remove(a.Content.Name())
}

func (a *attachment) setMetaData() error {
	defer a.Content.Seek(0, 0) // nolint: errcheck
	h := md5.New()
	size, err := io.Copy(h, a.Content)
	if err != nil {
		return err
	}
	a.Size = size
	a.Digest = fmt.Sprintf("md5-%x", h.Sum(nil))
	return nil
}

// copyDigest works the same as io.Copy, but also returns the md5sum of the
// copied file.
func copyDigest(tgt io.Writer, dst io.Reader) (int64, string, error) {
	h := md5.New()
	tee := io.TeeReader(dst, h)
	wg := sync.WaitGroup{}
	wg.Add(1)
	var written int64
	var err error
	go func() {
		written, err = io.Copy(tgt, tee)
		wg.Done()
	}()
	wg.Wait()
	return written, fmt.Sprintf("md5-%x", h.Sum(nil)), err
}

func (a *attachment) UnmarshalJSON(p []byte) error {
	var att struct {
		ContentType string `json:"content_type"`
		Stub        bool   `json:"stub"`
		Content     []byte `json:"data"`
		Size        int64  `json:"length"`
		Digest      string `json:"digest"`
	}
	if err := json.Unmarshal(p, &att); err != nil {
		return err
	}
	a.ContentType = att.ContentType
	a.Size = att.Size
	a.Digest = att.Digest
	if att.Stub {
		a.Stub = true
		return nil
	}
	tmp, err := ioutil.TempFile("", "attachment-")
	if err != nil {
		return err
	}
	size, digest, err := copyDigest(tmp, bytes.NewReader(att.Content))
	if err != nil {
		return err
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		return err
	}
	a.Content = tmp
	a.Size = size
	a.Digest = digest
	return nil
}

func (a *attachment) MarshalJSON() ([]byte, error) {
	if a.Content != nil {
		defer a.Content.Seek(0, 0) // nolint: errcheck
		if a.Size == 0 || a.Digest == "" {
			if err := a.setMetaData(); err != nil {
				return nil, err
			}
		}
	} else {
		if !a.Stub {
			return nil, errors.New("Content required")
		}
	}
	if a.Stub {
		return a.stubMarshalJSON()
	}

	type att struct {
		ContentType string `json:"content_type"`
		Data        []byte `json:"data"`
		Size        int64  `json:"length"`
		Digest      string `json:"digest"`
	}
	content, err := ioutil.ReadAll(a.Content)
	if err != nil {
		return nil, err
	}
	return json.Marshal(att{
		ContentType: a.ContentType,
		Data:        content,
		Size:        a.Size,
		Digest:      a.Digest,
	})
}

func (a *attachment) stubMarshalJSON() ([]byte, error) {
	type stub struct {
		ContentType string `json:"content_type"`
		Stub        bool   `json:"stub"`
		Size        int64  `json:"length,omitempty"`
		Digest      string `json:"digest,omitempty"`
	}
	return json.Marshal(stub{
		ContentType: a.ContentType,
		Stub:        true,
		Size:        a.Size,
		Digest:      a.Digest,
	})
}

type normalDoc struct {
	ID          string                 `json:"_id"`
	Rev         string                 `json:"_rev,omitempty"`
	Attachments attachments            `json:"_attachments,omitempty"`
	Data        map[string]interface{} `json:"-"`
}

func (d *normalDoc) MarshalJSON() ([]byte, error) {
	for _, key := range []string{"_id", "_rev", "_attachments"} {
		if _, ok := d.Data[key]; ok {
			return nil, errors.New("Data must not contain _id, _rev, or _attachments keys")
		}
	}
	var data []byte
	if len(d.Data) > 0 {
		var err error
		if data, err = json.Marshal(d.Data); err != nil {
			return nil, err
		}
	}
	doc, err := json.Marshal(*d)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		doc[len(doc)-1] = ','
		return append(doc, data[1:]...), nil
	}
	return doc, nil
}

func (d *normalDoc) UnmarshalJSON(p []byte) error {
	doc := struct {
		ID          string      `json:"_id"`
		Rev         string      `json:"_rev,omitempty"`
		Attachments attachments `json:"_attachments,omitempty"`
	}{}
	if err := json.Unmarshal(p, &doc); err != nil {
		return err
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(p, &data); err != nil {
		return err
	}
	delete(data, "_id")
	delete(data, "_rev")
	delete(data, "_attachments")
	d.ID = doc.ID
	d.Rev = doc.Rev
	d.Attachments = doc.Attachments
	d.Data = data
	return nil
}

func (d *normalDoc) cleanup() error {
	var err error
	for _, a := range d.Attachments {
		if e := a.cleanup(); e != nil {
			err = e
		}
	}
	return err
}

func normalizeDoc(i interface{}) (*normalDoc, error) {
	data, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	doc := &normalDoc{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, &kivik.Error{HTTPStatus: http.StatusBadRequest, Err: err}
	}
	return doc, nil
}