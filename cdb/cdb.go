// Package cdb provides the core CouchDB types.
package cdb

type Attachment struct{}

type RevHistory struct {
	Start int64    `json:"start" yaml:"start"`
	IDs   []string `json:"ids" yaml:"ids"`
}
