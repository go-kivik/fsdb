package fs

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/kivik"
	"gitlab.com/flimzy/testy"
)

func TestPut(t *testing.T) {
	type tt struct {
		fs       filesystem.Filesystem
		path     string
		dbname   string
		id       string
		doc      interface{}
		options  map[string]interface{}
		status   int
		err      string
		expected string
	}
	tests := testy.NewTable()
	tests.Add("invalid docID", tt{
		path:   "doesntmatter",
		dbname: "doesntmatter",
		id:     "_foo",
		status: http.StatusBadRequest,
		err:    "only reserved document ids may start with underscore",
	})
	tests.Add("invalid document", tt{
		path:   "doesntmatter",
		dbname: "doesntmatter",
		id:     "foo",
		doc:    make(chan int),
		status: http.StatusBadRequest,
		err:    "json: unsupported type: chan int",
	})
	tests.Add("create with revid", func(t *testing.T) interface{} {
		tmpdir := tempDir(t)
		tests.Cleanup(cleanTmpdir(tmpdir))
		if err := os.Mkdir(filepath.Join(tmpdir, "foo"), 0777); err != nil {
			t.Fatal(err)
		}

		return tt{
			path:   tmpdir,
			dbname: "foo",
			id:     "foo",
			doc:    map[string]string{"foo": "bar", "_rev": "1-xxx"},
			status: http.StatusConflict,
			err:    "document update conflict",
		}
	})
	tests.Add("simple create", func(t *testing.T) interface{} {
		tmpdir := tempDir(t)
		tests.Cleanup(cleanTmpdir(tmpdir))
		if err := os.Mkdir(filepath.Join(tmpdir, "foo"), 0777); err != nil {
			t.Fatal(err)
		}

		return tt{
			path:     tmpdir,
			dbname:   "foo",
			id:       "foo",
			doc:      map[string]string{"foo": "bar"},
			expected: "1-04edfaf9abdaed3c0accf6c463e78fd4",
		}
	})
	tests.Add("update conflict, doc key", func(t *testing.T) interface{} {
		tmpdir := copyDir(t, "testdata/db.put", 1)
		tests.Cleanup(cleanTmpdir(tmpdir))

		return tt{
			path:   tmpdir,
			dbname: "db.put",
			id:     "foo",
			doc:    map[string]string{"foo": "bar", "_rev": "2-asdf"},
			status: http.StatusConflict,
			err:    "document update conflict",
		}
	})
	tests.Add("update conflict, options", func(t *testing.T) interface{} {
		tmpdir := copyDir(t, "testdata/db.put", 1)
		tests.Cleanup(cleanTmpdir(tmpdir))

		return tt{
			path:    tmpdir,
			dbname:  "db.put",
			id:      "foo",
			doc:     map[string]string{"foo": "bar"},
			options: map[string]interface{}{"rev": "2-asdf"},
			status:  http.StatusConflict,
			err:     "document update conflict",
		}
	})
	tests.Add("no explicit rev", func(t *testing.T) interface{} {
		tmpdir := copyDir(t, "testdata/db.put", 1)
		tests.Cleanup(cleanTmpdir(tmpdir))

		return tt{
			path:   tmpdir,
			dbname: "db.put",
			id:     "foo",
			doc:    map[string]string{"foo": "bar"},
			status: http.StatusConflict,
			err:    "document update conflict",
		}
	})
	tests.Add("revs mismatch", tt{
		path:    "/tmp",
		dbname:  "doesntmatter",
		id:      "foo",
		doc:     map[string]string{"foo": "bar", "_rev": "2-asdf"},
		options: kivik.Options{"rev": "3-asdf"},
		status:  http.StatusBadRequest,
		err:     "document rev from request body and query string have different values",
	})
	tests.Add("proper update", func(t *testing.T) interface{} {
		tmpdir := copyDir(t, "testdata/db.put", 1)
		tests.Cleanup(cleanTmpdir(tmpdir))

		return tt{
			path:     tmpdir,
			dbname:   "db.put",
			id:       "foo",
			doc:      map[string]string{"foo": "quxx", "_rev": "1-beea34a62a215ab051862d1e5d93162e"},
			expected: "2-ff3a4f106331244679a6cac83a74ae48",
		}
	})
	tests.Add("design doc", func(t *testing.T) interface{} {
		tmpdir := tempDir(t)
		tests.Cleanup(cleanTmpdir(tmpdir))
		if err := os.Mkdir(filepath.Join(tmpdir, "foo"), 0777); err != nil {
			t.Fatal(err)
		}

		return tt{
			path:     tmpdir,
			dbname:   "foo",
			id:       "_design/foo",
			doc:      map[string]string{"foo": "bar"},
			expected: "1-04edfaf9abdaed3c0accf6c463e78fd4",
		}
	})
	tests.Add("invalid doc id", tt{
		path:   "/tmp",
		dbname: "doesntmatter",
		id:     "_oink",
		doc:    map[string]string{"foo": "bar"},
		status: http.StatusBadRequest,
		err:    "only reserved document ids may start with underscore",
	})
	tests.Add("invalid attachments", tt{
		path:   "/tmp",
		dbname: "doesntmatter",
		id:     "foo",
		doc: map[string]interface{}{
			"foo":          "bar",
			"_attachments": 123,
		},
		status: http.StatusBadRequest,
		err:    "json: cannot unmarshal number into Go struct field RevMeta._attachments of type map[string]*cdb.Attachment",
	})
	// tests.Add("attachment", func(t *testing.T) interface{} {
	// 	tmpdir := tempDir(t)
	// 	tests.Cleanup(cleanTmpdir(tmpdir))
	// 	if err := os.Mkdir(filepath.Join(tmpdir, "foo"), 0777); err != nil {
	// 		t.Fatal(err)
	// 	}
	//
	// 	return tt{
	// 		path:   tmpdir,
	// 		dbname: "foo",
	// 		id:     "foo",
	// 		doc: map[string]interface{}{
	// 			"foo": "bar",
	// 			"_attachments": map[string]interface{}{
	// 				"foo.txt": map[string]interface{}{
	// 					"content_type": "text/plain",
	// 					"data":         []byte("Testing"),
	// 				},
	// 			},
	// 		},
	// 		expected: "1-beea34a62a215ab051862d1e5d93162e",
	// 	}
	// })

	tests.Run(t, func(t *testing.T, tt tt) {
		if tt.path == "" {
			t.Fatalf("path must be set")
		}
		fs := tt.fs
		if fs == nil {
			fs = filesystem.Default()
		}
		c := &client{root: tt.path, fs: fs}
		db := c.newDB(tt.dbname)
		rev, err := db.Put(context.Background(), tt.id, tt.doc, tt.options)
		testy.StatusError(t, tt.err, tt.status, err)
		if rev != tt.expected {
			t.Errorf("Unexpected rev returned: %s", rev)
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), testy.JSONDir{
			Path:        tt.path,
			NoMD5Sum:    true,
			FileContent: true,
		}); d != nil {
			t.Error(d)
		}
	})
}
