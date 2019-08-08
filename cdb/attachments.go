package cdb

import (
	"bytes"
	"encoding/json"

	"github.com/go-kivik/fsdb/filesystem"
)

// Attachment represents a file attachment.
type Attachment struct {
	ContentType string `json:"content_type" yaml:"content_type"`
	RevPos      int64  `json:"revpos,omitempty" yaml:"revpos,omitempty"`
	Stub        bool   `json:"stub,omitempty" yaml:"stub,omitempty"`
	Follows     bool   `json:"follows,omitempty" yaml:"follows,omitempty"`
	Content     []byte `json:"data,omitempty" yaml:"content,omitempty"`
	Size        int64  `json:"length" yaml:"size"`
	Digest      string `json:"digest" yaml:"digest"`

	// path is the full path to the file on disk, or the empty string if the
	// attachment is not (yet) on disk.
	path string
	// fs is the filesystem to use for disk access.
	fs filesystem.Filesystem
}

// MarshalJSON implements the json.Marshaler interface.
func (a *Attachment) MarshalJSON() ([]byte, error) {
	if a.Stub || a.Follows {
		if err := a.readMetadata(); err != nil {
			return nil, err
		}
	} else {
		if err := a.readContent(); err != nil {
			return nil, err
		}
	}
	att := struct {
		Attachment
		Stub    *bool `json:"stub,omitempty"`
		Follows *bool `json:"follows,omitempty"`
	}{
		Attachment: *a,
	}
	if a.Stub {
		att.Stub = &a.Stub
	}
	if a.Follows {
		att.Follows = &a.Follows
	}
	return json.Marshal(att)
}

func (a *Attachment) readContent() error {
	f, err := a.fs.Open(a.path)
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	a.Size, a.Digest, err = copyDigest(buf, f)
	if err != nil {
		return err
	}
	a.Content = buf.Bytes()
	return nil
}

func (a *Attachment) readMetadata() error {
	f, err := a.fs.Open(a.path)
	if err != nil {
		return err
	}
	a.Size, a.Digest, err = digest(f)
	if err != nil {
		return err
	}
	return nil
}
