package fs

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestCompact(t *testing.T) {
	type tt struct {
		path   string
		dbname string
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("directory does not exist", tt{
		path:   "testdata",
		dbname: "notfound",
		status: http.StatusNotFound,
		err:    "^open testdata/notfound: no such file or directory$",
	})
	tests.Add("empty directory", func(t *testing.T) interface{} {
		tmpdir := tempDir(t)
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
		})
		if err := os.Mkdir(filepath.Join(tmpdir, "foo"), 0666); err != nil {
			t.Fatal(err)
		}

		return tt{
			path:   tmpdir,
			dbname: "foo",
		}
	})
	tests.Add("permission denied", func(t *testing.T) interface{} {
		tmpdir := tempDir(t)
		if err := os.Mkdir(filepath.Join(tmpdir, "foo"), 0); err != nil {
			t.Fatal(err)
		}
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
		})

		return tt{
			path:   tmpdir,
			dbname: "foo",
			status: http.StatusForbidden,
			err:    "/foo: permission denied$",
		}
	})
	tests.Add("abandoned attachments", func(t *testing.T) interface{} {
		tmpdir := copyDir(t, "testdata/compact.abandonedatt", 1)
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
		})

		return tt{
			path:   tmpdir,
			dbname: "compact.abandonedatt",
		}
	})
	tests.Add("no attachments", func(t *testing.T) interface{} {
		tmpdir := copyDir(t, "testdata/compact.noatt", 1)
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
		})

		return tt{
			path:   tmpdir,
			dbname: "compact.noatt",
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		db := &db{
			client: &client{root: tt.path},
			dbName: tt.dbname,
		}
		err := db.Compact(context.Background())
		testy.StatusErrorRE(t, tt.err, tt.status, err)
		if d := testy.DiffAsJSON(testy.Snapshot(t), testy.JSONDir{
			Path:        tt.path,
			NoMD5Sum:    true,
			FileContent: true,
		}); d != nil {
			t.Error(d)
		}
	})
}
