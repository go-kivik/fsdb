package cdb

import (
	"encoding/json"

	"github.com/go-kivik/kivik"
)

// Document is a CouchDB document.
type Document struct {
	ID        string      `json:"_id" yaml:"_id"`
	Revisions []*Revision `json:"-" yaml:"-"`

	Options kivik.Options `json:"-" yaml:"-"`
}

// MarshalJSON satisfies the json.Marshaler interface.
func (d *Document) MarshalJSON() ([]byte, error) {
	rev := d.Revisions[0]
	rev.options = d.Options
	revJSON, err := json.Marshal(rev)
	if err != nil {
		return nil, err
	}
	idJSON, _ := json.Marshal(map[string]string{
		"_id": d.ID,
	})
	return joinJSON(idJSON, revJSON), nil
}
