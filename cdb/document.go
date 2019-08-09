package cdb

import (
	"context"
	"encoding/json"

	"github.com/go-kivik/kivik"
)

// Document is a CouchDB document.
type Document struct {
	ID        string      `json:"_id" yaml:"_id"`
	Revisions []*Revision `json:"-" yaml:"-"`
	// RevsInfo is only used during JSON marshaling, and should never be
	// consulted as authoritative.
	RevsInfo []RevInfo `json:"_revs_info,omitempty" yaml:"-"`

	Options kivik.Options `json:"-" yaml:"-"`
}

// MarshalJSON satisfies the json.Marshaler interface.
func (d *Document) MarshalJSON() ([]byte, error) {
	d.revsInfo()
	rev := d.Revisions[0]
	rev.options = d.Options
	revJSON, err := json.Marshal(rev)
	if err != nil {
		return nil, err
	}
	docJSON, _ := json.Marshal(*d)
	return joinJSON(docJSON, revJSON), nil
}

// revsInfo populates the RevsInfo field, if appropriate according to options.
func (d *Document) revsInfo() {
	d.RevsInfo = nil
	if ok, _ := d.Options["revs_info"].(bool); !ok {
		return
	}
	if _, ok := d.Options["rev"]; ok {
		return
	}
	d.RevsInfo = make([]RevInfo, len(d.Revisions))
	for i, rev := range d.Revisions {
		d.RevsInfo[i] = RevInfo{
			Rev:    rev.Rev.String(),
			Status: "available",
		}
	}
}

// RevInfo is revisions information as presented in the _revs_info key.
type RevInfo struct {
	Rev    string `json:"rev"`
	Status string `json:"status"`
}

// Compact cleans up any non-leaf revs, and attempts to consolidate attachments.
func (d *Document) Compact(ctx context.Context) error {
	revTree := make(map[string][]*Revision, 1)
	// An index of ancestor -> leaf revision
	index := map[string]string{}
	for _, rev := range d.Revisions {
		revID := rev.Rev.String()
		if leaf, ok := index[revID]; ok {
			revTree[leaf] = append(revTree[leaf], rev)
			continue
		}
		for _, ancestor := range rev.RevHistory.ancestors()[1:] {
			index[ancestor] = revID
		}
		revTree[revID] = []*Revision{rev}
	}
	keep := make([]*Revision, 0, len(revTree))
	for _, rev := range d.Revisions {
		if _, ok := revTree[rev.Rev.String()]; ok {
			keep = append(keep, rev)
			continue
		}
		if err := rev.Delete(ctx); err != nil {
			return err
		}
	}
	d.Revisions = keep
	return nil
}
