// Package cdb provides the core CouchDB types.
package cdb

import (
	"fmt"
)

// RevHistory represents the recent ancestors of a revision.
type RevHistory struct {
	Start int64    `json:"start" yaml:"start"`
	IDs   []string `json:"ids" yaml:"ids"`
}

// ancestors returns the full, known revision history, newest first, starting
// with the current rev.
func (h *RevHistory) ancestors() []string {
	history := make([]string, len(h.IDs))
	for i, id := range h.IDs {
		history[i] = fmt.Sprintf("%d-%s", h.Start-int64(i), id)
	}
	return history
}
