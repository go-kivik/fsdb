package fs

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/go-kivik/fsdb/decoder"
	"github.com/go-kivik/fsdb/internal"
	"github.com/go-kivik/kivik"
)

func id2basename(id string) string {
	id = url.PathEscape(id)
	if id[0] == '.' {
		return "%2E" + id[1:]
	}
	return id
}

func filename2id(filename string) (string, error) {
	return url.PathUnescape(filename)
}

func (d *db) currentRev(docID, ext string) (string, error) {
	f, err := os.Open(d.path(docID))
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return "", err
	}
	r, err := decoder.Rev(f, ext)
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

func compareRevs(doc *internal.Document, opts map[string]interface{}, currev string) error {
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
func (d *db) Put(_ context.Context, docID string, i interface{}, opts map[string]interface{}) (string, error) {
	if err := validateID(docID); err != nil {
		return "", err
	}
	rev, err := d.cdb.NewRevision(i)
	if err != nil {
		return "", err
	}
	doc, err := d.cdb.OpenDocID(docID, opts)
	switch {
	case kivik.StatusCode(err) == http.StatusNotFound:
		// Crate new doc
		doc = d.cdb.NewDocument(docID)
	case err != nil:
		return "", err
	}
	return doc.AddRevision(rev, opts)

	// filename := id2basename(docID) + ".json"
	// ndoc, err := normalizeDoc(doc)
	// if err != nil {
	// 	return "", err
	// }
	// defer ndoc.Cleanup() // nolint: errcheck
	// currev, err := d.currentRev(filename, "json")
	// if err != nil {
	// 	return "", err
	// }
	// atts := ndoc.Attachments
	//
	// if err := compareRevs(ndoc, opts, currev); err != nil {
	// 	return "", err
	// }
	// ndoc.ID = docID
	// ndoc.Rev.Increment(fmt.Sprintf("%d %s", ndoc.Rev.Seq+1, ndoc.ID))
	//
	// // map of tmpFile:permFile to be renamed
	// toRename := make(map[string]string)
	// defer func() {
	// 	for filename := range toRename {
	// 		_ = os.Remove(filename)
	// 	}
	// }()
	//
	// if atts != nil {
	// 	base := base(filename)
	// 	if err := os.Mkdir(d.path(base), 0777); err != nil {
	// 		return "", err
	// 	}
	// 	for attname, att := range atts {
	// 		tmp, err := ioutil.TempFile(d.path(base), ".")
	// 		if err != nil {
	// 			return "", err
	// 		}
	// 		toRename[tmp.Name()] = d.path(base, attname)
	// 		if _, err := io.Copy(tmp, att.Content); err != nil {
	// 			return "", err
	// 		}
	// 		att.Stub = true
	// 	}
	// }
	//
	// tmp, err := ioutil.TempFile(d.path(), ".")
	// if err != nil {
	// 	return "", err
	// }
	// defer tmp.Close() // nolint: errcheck
	// toRename[tmp.Name()] = d.path(filename)
	// if err := json.NewEncoder(tmp).Encode(ndoc); err != nil {
	// 	return "", err
	// }
	//
	// if err := tmp.Close(); err != nil {
	// 	return "", err
	// }
	// if currev != "" {
	// 	if err := d.archiveDoc(filename, currev); err != nil {
	// 		return "", err
	// 	}
	// }
	// for old, new := range toRename {
	// 	if err := os.Rename(old, new); err != nil {
	// 		return "", err
	// 	}
	// }
	// return ndoc.Rev.String(), nil
}
