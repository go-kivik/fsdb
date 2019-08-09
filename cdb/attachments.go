package cdb

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/kivik/driver"
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

// Open opens the attachment for reading.
func (a *Attachment) Open() (filesystem.File, error) {
	if a.path == "" {
		return nil, errors.New("no path defined")
	}
	return a.fs.Open(a.path)
}

// MarshalJSON implements the json.Marshaler interface.
func (a *Attachment) MarshalJSON() ([]byte, error) {
	var err error
	switch {
	case len(a.Content) != 0:
		err = a.setMetadata()
	case a.Stub || a.Follows:
		err = a.readMetadata()
	default:
		err = a.readContent()
	}
	if err != nil {
		return nil, err
	}
	att := struct {
		Attachment
		Stub    *bool `json:"stub,omitempty"`    // nolint: govet
		Follows *bool `json:"follows,omitempty"` // nolint: govet
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
	return err
}

func (a *Attachment) setMetadata() error {
	a.Size = int64(len(a.Content))
	var err error
	_, a.Digest, err = digest(bytes.NewReader(a.Content))
	return err
}

type attsIter []*driver.Attachment

var _ driver.Attachments = &attsIter{}

func (i attsIter) Close() error {
	for _, att := range i {
		if err := att.Content.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (i *attsIter) Next(att *driver.Attachment) error {
	if len(*i) == 0 {
		return io.EOF
	}
	var next *driver.Attachment
	next, *i = (*i)[0], (*i)[1:]
	*att = *next
	return nil
}

// AttachmentsIterator will return a driver.Attachments iterator, if the options
// permit. If options don't permit, both return values will be nil.
func (r *Revision) AttachmentsIterator() (driver.Attachments, error) {
	if attachments, _ := r.options["attachments"].(bool); !attachments {
		return nil, nil
	}
	if accept, _ := r.options["header:accept"].(string); accept == "application/json" {
		return nil, nil
	}
	iter := make(attsIter, 0, len(r.Attachments))
	for filename, att := range r.Attachments {
		f, err := att.Open()
		if err != nil {
			return nil, err
		}
		iter = append(iter, &driver.Attachment{
			Filename:    filename,
			Content:     f,
			ContentType: att.ContentType,
			Stub:        att.Stub,
			Follows:     att.Follows,
			Size:        att.Size,
			RevPos:      att.RevPos,
			Digest:      att.Digest,
		})
	}
	return &iter, nil
}
