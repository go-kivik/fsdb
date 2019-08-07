package cdb

import "net/url"

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

func escapeID(id string) string {
	if id == "" {
		return id
	}
	id = url.PathEscape(id)
	if id[0] == '.' {
		return "%2E" + id[1:]
	}
	return id
}

func unescapeID(filename string) (string, error) {
	return url.PathUnescape(filename)
}
