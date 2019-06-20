package fs

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type attachments map[string]*attachment

type attachment struct {
	ContentType string   `json:"content_type"`
	Stub        bool     `json:"stub,omitempty"`
	Content     *os.File `json:"data,omitempty"`
	Size        int64    `json:"size"`
	Digest      string   `json:"digest"`
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
