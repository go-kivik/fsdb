package cdb

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/go-kivik/fsdb/filesystem"
	"gitlab.com/flimzy/testy"
)

func TestDocumentPersist(t *testing.T) {
	type tt struct {
		path   string
		doc    *Document
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("nil doc", func(t *testing.T) interface{} {
		var tmpdir string
		tests.Cleanup(testy.TempDir(t, &tmpdir))

		return tt{
			path:   tmpdir,
			status: http.StatusBadRequest,
			err:    "document has no revisions",
		}
	})
	tests.Add("no revs", func(t *testing.T) interface{} {
		var tmpdir string
		tests.Cleanup(testy.TempDir(t, &tmpdir))

		cdb := New(tmpdir, filesystem.Default())

		return tt{
			path:   tmpdir,
			doc:    cdb.NewDocument(tmpdir, "foo"),
			status: http.StatusBadRequest,
			err:    "document has no revisions",
		}
	})
	tests.Add("new doc, one rev", func(t *testing.T) interface{} {
		var tmpdir string
		tests.Cleanup(testy.TempDir(t, &tmpdir))

		cdb := New(tmpdir, filesystem.Default())
		doc := cdb.NewDocument(tmpdir, "foo")
		rev, _ := cdb.NewRevision(map[string]string{
			"_rev":  "1-xxx",
			"value": "bar",
		})
		doc.AddRevision(rev, nil)

		return tt{
			path: tmpdir,
			doc:  doc,
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		err := tt.doc.persist()
		testy.StatusError(t, tt.err, tt.status, err)
		re := testy.Replacement{
			Regexp:      regexp.MustCompile(regexp.QuoteMeta(tt.path)),
			Replacement: "<tmpdir>",
		}
		if d := testy.DiffInterface(testy.Snapshot(t, "doc"), tt.doc, re); d != nil {
			t.Error(d)
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t, "fs"), testy.JSONDir{
			Path:        tt.path,
			NoMD5Sum:    true,
			FileContent: true,
		}); d != nil {
			t.Error(d)
		}
	})
}
