package internal

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/kivik/driver"
)

type Attachments map[string]*Attachment

var _ driver.Attachments = Attachments{}

func (a Attachments) Close() error {
	for filename, att := range a {
		if err := att.Content.Close(); err != nil {
			return err
		}
		delete(a, filename)
	}
	return nil
}

func (a Attachments) Next(driverAtt *driver.Attachment) error {
	for filename, att := range a {
		x := &driver.Attachment{
			Filename:    filename,
			ContentType: att.ContentType,
			Stub:        att.Stub,
			Content:     att.Content,
			Size:        att.Size,
			Digest:      att.Digest,
			RevPos:      att.RevPos,
		}
		*driverAtt = *x
		delete(a, filename)
		return nil
	}
	return io.EOF
}

// Attachment represents a file attachment.
type Attachment struct {
	ContentType string          `json:"content_type"`
	RevPos      int64           `json:"revpos,omitempty"`
	Stub        bool            `json:"stub,omitempty"`
	Follows     bool            `json:"follows,omitempty"`
	Content     filesystem.File `json:"data,omitempty"`
	Size        int64           `json:"length"`
	Digest      string          `json:"digest"`
}

func (a *Attachment) Cleanup() error {
	if a == nil || a.Content == nil {
		return nil
	}
	if err := a.Content.Close(); err != nil {
		return err
	}
	return os.Remove(a.Content.Name())
}

func (a *Attachment) setMetaData() error {
	defer a.Content.Seek(0, 0) // nolint: errcheck
	h := md5.New()
	size, err := io.Copy(h, a.Content)
	if err != nil {
		return err
	}
	a.Size = size
	a.Digest = fmt.Sprintf("md5-%s", base64.StdEncoding.EncodeToString(h.Sum(nil)))
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

func (a *Attachment) UnmarshalJSON(p []byte) error {
	var att struct {
		ContentType string `json:"content_type"`
		RevPos      int64  `json:"revpos"`
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
	a.RevPos = att.RevPos
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

func (a *Attachment) MarshalJSON() ([]byte, error) {
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
	if a.Follows {
		return a.followsMarshalJSON()
	}

	type att struct {
		ContentType string `json:"content_type"`
		RevPos      int64  `json:"revpos,omitempty"`
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
		RevPos:      a.RevPos,
		Data:        content,
		Size:        a.Size,
		Digest:      a.Digest,
	})
}

func (a *Attachment) stubMarshalJSON() ([]byte, error) {
	type stub struct {
		ContentType string `json:"content_type"`
		RevPos      int64  `json:"revpos,omitempty"`
		Stub        bool   `json:"stub"`
		Size        int64  `json:"length,omitempty"`
		Digest      string `json:"digest,omitempty"`
	}
	return json.Marshal(stub{
		ContentType: a.ContentType,
		RevPos:      a.RevPos,
		Stub:        true,
		Size:        a.Size,
		Digest:      a.Digest,
	})
}

func (a *Attachment) followsMarshalJSON() ([]byte, error) {
	type stub struct {
		ContentType string `json:"content_type"`
		RevPos      int64  `json:"revpos,omitempty"`
		Follows     bool   `json:"follows"`
		Size        int64  `json:"length,omitempty"`
		Digest      string `json:"digest,omitempty"`
	}
	return json.Marshal(stub{
		ContentType: a.ContentType,
		RevPos:      a.RevPos,
		Follows:     true,
		Size:        a.Size,
		Digest:      a.Digest,
	})
}
