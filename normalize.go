package fs

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

type attachments map[string]*attachment

type attachment struct {
	ContentType string   `json:"content_type"`
	Stub        bool     `json:"stub,omitempty"`
	Content     *os.File `json:"data,omitempty"`
	Size        int64    `json:"size"`
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
		Size        int64  `json:"size"`
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
	if a.Content == nil {
		return nil, errors.New("Content required")
	}
	defer a.Content.Seek(0, 0) // nolint: errcheck
	if a.Size == 0 || a.Digest == "" {
		if err := a.setMetaData(); err != nil {
			return nil, err
		}
	}
	if a.Stub {
		return a.stubMarshalJSON()
	}

	type att struct {
		ContentType string `json:"content_type"`
		Data        []byte `json:"data"`
		Size        int64  `json:"size"`
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
		Size        int64  `json:"size,omitempty"`
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
	Rev         string                 `json:"_rev"`
	Attachments attachments            `json:"_attachments"`
	Data        map[string]interface{} `json:"-"`
}
