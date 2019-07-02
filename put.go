package fs

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/go-kivik/kivik"
)

func id2filename(id string) string {
	id = url.PathEscape(id)
	if id[0] == '.' {
		return "%2E" + id[1:] + ".json"
	}
	return id + ".json"
}

type revDoc struct {
	Rev rev `json:"_rev"`
}

type rev struct {
	seq int64
	sum string
}

func (r *rev) UnmarshalJSON(p []byte) error {
	if p[0] == '"' {
		var str string
		if e := json.Unmarshal(p, &str); e != nil {
			return e
		}
		parts := strings.SplitN(str, "-", 2)
		seq, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return err
		}
		r.seq = seq
		if len(parts) > 1 {
			r.sum = parts[1]
		}
		return nil
	}
	r.sum = ""
	return json.Unmarshal(p, &r.sum)
}

func (r rev) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r rev) String() string {
	if r.seq == 0 {
		return ""
	}
	return fmt.Sprintf("%d-%s", r.seq, r.sum)
}

func (r rev) IsZero() bool {
	return r.seq == 0
}

func (r *rev) Increment(payload string) {
	r.seq++
	if payload == "" {
		r.sum = ""
		return
	}
	r.sum = fmt.Sprintf("%x", md5.Sum([]byte(payload)))
}

func (d *db) currentRev(docID string) (string, error) {
	f, err := os.Open(d.path(docID))
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return "", err
	}
	var rd revDoc
	if e := json.NewDecoder(f).Decode(&rd); e != nil {
		return "", e
	}
	return rd.Rev.String(), nil
}

func compareRevs(doc *normalDoc, opts map[string]interface{}, currev string) error {
	optsrev, _ := opts["rev"].(string)
	docrev := doc.Rev.String()
	if optsrev != "" && docrev != "" && optsrev != docrev {
		return &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "document rev from request body and query string have different values"}
	}
	if currev == "" {
		return nil
	}
	newrev := optsrev
	if newrev == "" {
		newrev = docrev
	}
	if newrev != currev {
		return &kivik.Error{HTTPStatus: http.StatusConflict, Message: "document update conflict"}
	}
	return nil
}

func base(p string) string {
	return strings.TrimSuffix(p, path.Ext(p))
}

func (d *db) archiveDoc(filename, rev string) error {
	ext := path.Ext(filename)
	base := base(filename)
	src, err := os.Open(d.path(filename))
	if err != nil {
		return err
	}
	defer src.Close() // nolint: errcheck
	if err := os.Mkdir(d.path("."+base), 0777); err != nil {
		return err
	}
	tmp, err := ioutil.TempFile(d.path("."+base), ".")
	if err != nil {
		return err
	}
	defer tmp.Close() // nolint: errcheck
	if _, err := io.Copy(tmp, src); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), d.path("."+base, rev)+ext)
}

var reservedPrefixes = []string{"_local/", "_design/"}

func validateID(id string) error {
	if id[0] != '_' {
		return nil
	}
	for _, prefix := range reservedPrefixes {
		if strings.HasPrefix(id, prefix) && len(id) > len(prefix) {
			return nil
		}
	}
	return &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "only reserved document ids may start with underscore"}
}

/*
File naming strategy:
Current rev lives under:    {db}/{docid}.{ext}
Historical revs live under: {db}/.{docid}/{rev}
Attachments:                {db}/{docid}/{filename}
*/
func (d *db) Put(_ context.Context, docID string, doc interface{}, opts map[string]interface{}) (string, error) {
	if err := validateID(docID); err != nil {
		return "", err
	}
	filename := id2filename(docID)
	ndoc, err := normalizeDoc(doc)
	if err != nil {
		return "", err
	}
	defer ndoc.cleanup() // nolint: errcheck
	currev, err := d.currentRev(filename)
	if err != nil {
		return "", err
	}
	atts := ndoc.Attachments

	if err := compareRevs(ndoc, opts, currev); err != nil {
		return "", err
	}
	ndoc.ID = docID
	ndoc.Rev.Increment(fmt.Sprintf("%d %s", ndoc.Rev.seq+1, ndoc.ID))

	// map of tmpFile:permFile to be renamed
	toRename := make(map[string]string)
	defer func() {
		for filename := range toRename {
			_ = os.Remove(filename)
		}
	}()

	if atts != nil {
		base := base(filename)
		if err := os.Mkdir(d.path(base), 0777); err != nil {
			return "", err
		}
		for attname, att := range atts {
			tmp, err := ioutil.TempFile(d.path(base), ".")
			if err != nil {
				return "", err
			}
			toRename[tmp.Name()] = d.path(base, attname)
			if _, err := io.Copy(tmp, att.Content); err != nil {
				return "", err
			}
			att.Stub = true
		}
	}

	tmp, err := ioutil.TempFile(d.path(), ".")
	if err != nil {
		return "", err
	}
	defer tmp.Close() // nolint: errcheck
	toRename[tmp.Name()] = d.path(filename)
	if err := json.NewEncoder(tmp).Encode(ndoc); err != nil {
		return "", err
	}

	if err := tmp.Close(); err != nil {
		return "", err
	}
	if currev != "" {
		if err := d.archiveDoc(filename, currev); err != nil {
			return "", err
		}
	}
	for old, new := range toRename {
		if err := os.Rename(old, new); err != nil {
			return "", err
		}
	}
	return ndoc.Rev.String(), nil
}
