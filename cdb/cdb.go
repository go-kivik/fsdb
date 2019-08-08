// Package cdb provides the core CouchDB types.
package cdb

// RevHistory represents the recent ancestors of a revision.
type RevHistory struct {
	Start int64    `json:"start" yaml:"start"`
	IDs   []string `json:"ids" yaml:"ids"`
}
