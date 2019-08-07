package cdb

import (
	"net/http"
	"testing"

	"github.com/go-kivik/fsdb/filesystem"
	"gitlab.com/flimzy/testy"
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

	tests.Run(t, func(t *testing.T, tt tt) {
		fs := New(tt.root, tt.fs)
		result, err := fs.Open(tt.docID)
		testy.StatusError(t, tt.err, tt.status, err)
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}
