package internal

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/xerrors"
)

const (
	// Maybe make this confiurable at some point?
	revsLimit = 1000
)

// RevsInfo is revisions information as presented in the _revs_info key.
type RevsInfo struct {
	Rev    string `json:"rev"`
	Status string `json:"status"`
}

// Revisions is historical revisions data, as presented in the _revisions key.
type Revisions struct {
	Start int64    `json:"start"`
	IDs   []string `json:"ids"`
}

// DocMeta contains the special CouchDB metadata fields for each document.
type DocMeta struct {
	ID          string      `json:"_id,omitempty" yaml:"_id,omitempty"`
	Rev         RevID       `json:"_rev,omitempty" yaml:"_rev,omitempty"`
	Attachments Attachments `json:"_attachments,omitempty" yaml:"_attachments,omitempty"`
	RevsInfo    []RevsInfo  `json:"_revs_info,omitempty" yaml:"_revs_info,omitempty"`
	Revisions   *Revisions  `json:"_revisions,omitempty" yaml:"_revisions,omitempty"`

	// Path is the full, on-disk path to this document.
	Path string `json:"-" yaml:"-"`
}

// DocMetas is a list of document metadata, which can be sorted using CouchDB's
// conflict resolution logic.
type DocMetas []*DocMeta

var _ sort.Interface = DocMetas{}

// Len returns the number of revisions in r.
func (r DocMetas) Len() int {
	return len(r)
}

// Less allows sorting r.
func (r DocMetas) Less(i, j int) bool {
	return r[i].Rev.Seq > r[j].Rev.Seq ||
		(r[i].Rev.Seq == r[j].Rev.Seq && r[i].Rev.Sum > r[j].Rev.Sum)
}

func (r DocMetas) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// Document is a CouchDB document.
type Document struct {
	DocMeta
	Data map[string]interface{} `json:"-"`
	Path string                 `json:"-"`
}

// MarshalJSON satisfies the json.Marshaler interface.
func (d *Document) MarshalJSON() ([]byte, error) {
	for key := range d.Data {
		if key[0] == '_' {
			return nil, xerrors.Errorf("Bad special document member: %s", key)
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

var reservedKeys = map[string]struct{}{
	"_id":                {},
	"_rev":               {},
	"_attachments":       {},
	"_revisions":         {},
	"_revs_info":         {}, // *
	"_deleted":           {},
	"_conflicts":         {}, // *
	"_deleted_conflicts": {}, // *
	"_local_seq":         {}, // *
	// * Can these be PUT?
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (d *Document) UnmarshalJSON(p []byte) error {
	if err := json.Unmarshal(p, &d.DocMeta); err != nil {
		return err
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(p, &data); err != nil {
		return err
	}
	for key := range data {
		if key[0] == '_' {
			if _, ok := reservedKeys[key]; !ok {
				return xerrors.Errorf("Bad special document member: %s", key)
			}
			delete(data, key)
		}
	}
	d.Data = data
	return nil
}

func (d *Document) Cleanup() error {
	var err error
	for _, a := range d.Attachments {
		if e := a.Cleanup(); e != nil {
			err = e
		}
	}
	return err
}

func (d *Document) GetRevsInfo() []RevsInfo {
	// This driver only ever stores the current leaf, so we only ever
	// return one revision here: the current one.
	return []RevsInfo{
		{
			Rev:    d.Rev.String(),
			Status: "available",
		},
	}
}

func (d *Document) GetRevisions() *Revisions {
	if d.Revisions != nil {
		return d.Revisions
	}
	histSize := d.Rev.Seq
	if histSize > revsLimit {
		histSize = revsLimit
	}
	var ids []string
	if d.Rev.Sum == "" {
		ids = make([]string, int(histSize))
	} else {
		ids = []string{d.Rev.Sum}
	}
	return &Revisions{
		Start: d.Rev.Seq,
		IDs:   ids,
	}
}

// Ext returns d.Path's file extension, without a dot. If d.Path is empty, Ext()
// returns the empty string.
func (d *DocMeta) Ext() string {
	return strings.TrimPrefix(filepath.Ext(d.Path), ".")
}