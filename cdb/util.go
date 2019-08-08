package cdb

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/url"
	"sync"
)

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

// copyDigest works the same as io.Copy, but also returns the md5sum digest of
// the copied file.
func copyDigest(tgt io.Writer, dst io.Reader) (int64, string, error) {
	h := md5.New()
	tee := io.TeeReader(dst, h)
	wg := sync.WaitGroup{}
	wg.Add(1)
	var written int64
	var err error
	go func() {
		written, err = io.Copy(tgt, tee)
		wg.Done()
	}()
	wg.Wait()

	return written, "md5-" + base64.StdEncoding.EncodeToString(h.Sum(nil)), err
}

func joinJSON(objects ...json.RawMessage) []byte {
	var size int
	for _, obj := range objects {
		size += len(obj)
	}
	result := make([]byte, 0, size)
	result = append(result, '{')
	for _, obj := range objects {
		if len(obj) == 4 && string(obj) == "null" {
			continue
		}
		result = append(result, obj[1:len(obj)-1]...)
		result = append(result, ',')
	}
	result[len(result)-1] = '}'
	return result
}
