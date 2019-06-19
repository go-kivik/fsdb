package fs

import (
	"context"
	"net/http"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"
)

func TestPut(t *testing.T) {
	type tst struct {
		setup    func(*testing.T, *db)
		final    func(*testing.T, *db)
		id       string
		doc      interface{}
		options  map[string]interface{}
		status   int
		err      string
		expected string
	}
	tests := testy.NewTable()
	tests.Add("simple create", tst{
		id:       "foo",
		doc:      map[string]string{"foo": "bar"},
		expected: "1-beea34a62a215ab051862d1e5d93162e",
		final: func(t *testing.T, d *db) {
			expected := map[string]string{
				"_rev": "1-beea34a62a215ab051862d1e5d93162e",
				"_id":  "foo",
				"foo":  "bar",
			}
			if d := diff.AsJSON(expected, &diff.File{Path: d.path("foo")}); d != nil {
				t.Error(d)
			}
		},
	})
	tests.Add("update conflict, doc key", tst{
		id:  "foo",
		doc: map[string]string{"foo": "bar", "_rev": "2-asdf"},
		setup: func(t *testing.T, d *db) {
			if _, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, nil); err != nil {
				t.Fatal(err)
			}
		},
		status: http.StatusConflict,
		err:    "document update conflict",
	})
	tests.Add("update conflict, options", tst{
		id:  "foo",
		doc: map[string]string{"foo": "bar"},
		setup: func(t *testing.T, d *db) {
			if _, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, nil); err != nil {
				t.Fatal(err)
			}
		},
		options: map[string]interface{}{"rev": "2-asdf"},
		status:  http.StatusConflict,
		err:     "document update conflict",
	})
	tests.Add("revs mismatch", tst{
		id:  "foo",
		doc: map[string]string{"foo": "bar", "_rev": "2-asdf"},
		setup: func(t *testing.T, d *db) {
			if _, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, nil); err != nil {
				t.Fatal(err)
			}
		},
		options: map[string]interface{}{"rev": "3-asdf"},
		status:  http.StatusBadRequest,
		err:     "document rev from request body and query string have different values",
	})
	tests.Add("new rev", tst{
		id:       "foo",
		doc:      map[string]string{"foo": "quxx", "_rev": "1-beea34a62a215ab051862d1e5d93162e"},
		expected: "2-a1de8ffe0af07dec9193ddf8d4b18135",
		setup: func(t *testing.T, d *db) {
			if _, err := d.Put(context.Background(), "foo", map[string]string{"foo": "bar"}, nil); err != nil {
				t.Fatal(err)
			}
		},
		final: func(t *testing.T, d *db) {
			expected := map[string]string{
				"_rev": "2-a1de8ffe0af07dec9193ddf8d4b18135",
				"_id":  "foo",
				"foo":  "quxx",
			}
			if d := diff.AsJSON(expected, &diff.File{Path: d.path("foo")}); d != nil {
				t.Error(d)
			}
			expected2 := map[string]string{
				"foo":  "bar",
				"_id":  "foo",
				"_rev": "1-beea34a62a215ab051862d1e5d93162e",
			}
			if d := diff.AsJSON(expected2, &diff.File{Path: d.path(".foo", "1-beea34a62a215ab051862d1e5d93162e")}); d != nil {
				t.Error(d)
			}
		},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		tmpdir := tempDir(t)
		defer rmdir(t, tmpdir)
		db := &db{
			client: &client{root: tmpdir},
		}
		if test.setup != nil {
			test.setup(t, db)
		}
		rev, err := db.Put(context.Background(), test.id, test.doc, test.options)
		testy.StatusError(t, test.err, test.status, err)
		if rev != test.expected {
			t.Errorf("Unexpected rev returned: %s", rev)
		}
		if test.final != nil {
			test.final(t, db)
		}
	})
}
