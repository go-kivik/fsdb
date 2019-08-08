package cdb

import "encoding/json"

// Document is a CouchDB document.
type Document struct {
	ID        string      `json:"_id" yaml:"_id"`
	Revisions []*Revision `json:"-" yaml:"-"`
}

// MarshalJSON satisfies the json.Marshaler interface.
func (d *Document) MarshalJSON() ([]byte, error) {
	revJSON, err := json.Marshal(d.Revisions[0])
	if err != nil {
		return nil, err
	}
	idJSON, _ := json.Marshal(map[string]string{
		"_id": d.ID,
	})
	return joinJSON(idJSON, revJSON), nil
}
