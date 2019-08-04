package fs

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/kivik"
)

func TestCompact(t *testing.T) {
	type tt struct {
		fs     filesystem.Filesystem
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
	tests.Add("permission denied", tt{
		fs: &filesystem.MockFS{
			OpenFunc: func(_ string) (filesystem.File, error) {
				return nil, &kivik.Error{HTTPStatus: http.StatusForbidden, Err: errors.New("permission denied")}
			},
		},
		path:   "somepath",
		dbname: "doesntmatter",
		status: http.StatusForbidden,
		err:    "permission denied$",
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
	tests.Add("non-winning revs only, no attachments", func(t *testing.T) interface{} {
		tmpdir := copyDir(t, "testdata/compact.nowinner_noatt", 1)
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
		})

		return tt{
			path:   tmpdir,
			dbname: "compact.nowinner_noatt",
		}
	})
	tests.Add("non-winning, abandoned attachments", func(t *testing.T) interface{} {
		tmpdir := copyDir(t, "testdata/compact.nowinner_abandonedatt", 1)
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
		})

		return tt{
			path:   tmpdir,
			dbname: "compact.nowinner_abandonedatt",
		}
	})
	tests.Add("split attachments", func(t *testing.T) interface{} {
		// Some attachments stored with winning rev, some stored
		// in revs dir. This simulates an aborted update.
		tmpdir := copyDir(t, "testdata/compact.split_atts", 1)
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
		})

		return tt{
			path:   tmpdir,
			dbname: "compact.split_atts",
		}
	})
	tests.Add("extra attachments", func(t *testing.T) interface{} {
		// Some attachments stored with winning rev, some stored
		// in revs dir. This simulates an aborted update.
		tmpdir := copyDir(t, "testdata/compact.extra_atts", 1)
		tests.Cleanup(func() error {
			return os.RemoveAll(tmpdir)
		})

		return tt{
			path:   tmpdir,
			dbname: "compact.extra_atts",
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		fs := tt.fs
		if fs == nil {
			fs = filesystem.Default()
		}
		db := &db{
			client: &client{root: tt.path},
			dbName: tt.dbname,
		}
		err := db.compact(context.Background(), fs)
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
