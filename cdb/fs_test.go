package cdb

import (
	"errors"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/kivik"
)

func TestFSOpen(t *testing.T) {
	type tt struct {
		fs     filesystem.Filesystem
		root   string
		docID  string
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("not found", tt{
		root:   "notfound",
		status: http.StatusNotFound,
		err:    "missing",
	})
	tests.Add("main rev only", tt{
		root:  "testdata/open",
		docID: "foo",
	})
	tests.Add("main rev only, yaml", tt{
		root:  "testdata/open",
		docID: "bar",
	})
	tests.Add("no id in doc", tt{
		root:  "testdata/open",
		docID: "noid",
	})
	tests.Add("forbidden", tt{
		fs: &filesystem.MockFS{
			OpenFunc: func(_ string) (filesystem.File, error) {
				return nil, &kivik.Error{HTTPStatus: http.StatusForbidden, Err: errors.New("permission denied")}
			},
		},
		root:   "doesntmatter",
		docID:  "foo",
		status: http.StatusForbidden,
		err:    "permission denied",
	})
	tests.Add("attachment", tt{
		root:  "testdata/open",
		docID: "att",
	})
	tests.Add("attachments from multiple revs", tt{
		root:  "testdata/open",
		docID: "splitatts",
	})
	tests.Add("no rev", tt{
		root:  "testdata/open",
		docID: "norev",
	})
	tests.Add("no main rev", tt{
		root:  "testdata/open",
		docID: "nomain",
	})
	tests.Add("json auto rev number", tt{
		root:  "testdata/open",
		docID: "jsonautorevnum",
	})
	tests.Add("yaml auto rev number", tt{
		root:  "testdata/open",
		docID: "yamlautorevnum",
	})
	tests.Add("json auto rev string", tt{
		root:  "testdata/open",
		docID: "jsonautorevstr",
	})
	tests.Add("yaml auto rev string", tt{
		root:  "testdata/open",
		docID: "yamlautorevstr",
	})
	tests.Add("multiple revs, winner selected", tt{
		root:  "testdata/open",
		docID: "multiplerevs",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		fs := New(tt.root, tt.fs)
		result, err := fs.Open(tt.docID)
		testy.StatusError(t, tt.err, tt.status, err)
		result.Options = kivik.Options{
			"_revs": true,
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}
