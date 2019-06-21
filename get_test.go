package fs

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"
	"github.com/go-kivik/kivik/driver"
)

func TestGet(t *testing.T) {
	type tt struct {
		setup    func(*testing.T, *db)
		final    func(*testing.T, *db)
		path     string
		id       string
		options  map[string]interface{}
		expected *driver.Document
		status   int
		err      string
	}
	tests := testy.NewTable()
	tests.Add("no id", tt{
		status: http.StatusBadRequest,
		err:    "no docid specified",
	})
	tests.Add("not found", tt{
		id:     "foo",
		status: http.StatusNotFound,
		err:    `/foo.json: no such file or directory$`,
	})
	tests.Add("forbidden", func(t *testing.T) interface{} {
		var mode os.FileMode
		return tt{
			setup: func(t *testing.T, d *db) {
				f, err := os.Open(d.path())
				if err != nil {
					t.Fatal(err)
				}
				stat, err := f.Stat()
				if err != nil {
					t.Fatal(err)
				}
				mode = stat.Mode()
				if err := os.Chmod(d.path(), 0000); err != nil {
					t.Fatal(err)
				}
			},
			final: func(t *testing.T, d *db) {
				if err := os.Chmod(d.path(), mode); err != nil {
					t.Fatal(err)
				}
			},
			id:     "foo",
			status: http.StatusForbidden,
			err:    "/foo.json: permission denied$",
		}
	})
	tests.Add("success, no attachments", tt{
		path: "testdata/db.foo",
		id:   "noattach",
		expected: &driver.Document{
			ContentLength: 72,
			Rev:           "1-xxxxxxxxxx",
		},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		dir := tt.path
		if dir == "" {
			dir = tempDir(t)
			defer rmdir(t, dir)
		}
		db := &db{
			client: &client{root: dir},
		}
		if tt.setup != nil {
			tt.setup(t, db)
		}
		doc, err := db.Get(context.Background(), tt.id, tt.options)
		testy.StatusErrorRE(t, tt.err, tt.status, err)
		defer doc.Body.Close() // nolint: errcheck
		if d := diff.AsJSON(&diff.File{Path: "testdata/" + testy.Stub(t)}, doc.Body); d != nil {
			t.Errorf("document:\n%s", d)
		}
		doc.Body = nil
		if d := diff.Interface(tt.expected, doc); d != nil {
			t.Error(d)
		}
		if tt.final != nil {
			tt.final(t, db)
		}
	})
}
