package fs

import (
	"context"
	"testing"

	"github.com/go-kivik/fsdb/v3/filesystem"
	"gitlab.com/flimzy/testy"
)

func TestSecurity(t *testing.T) {
	type tt struct {
		fs           filesystem.Filesystem
		path, dbname string
		status       int
		err          string
	}
	tests := testy.NewTable()
	tests.Add("no security object", tt{
		dbname: "foo",
	})
	tests.Add("json security obj", tt{
		path:   "testdata",
		dbname: "db_foo",
	})
	tests.Add("yaml security obj", tt{
		path:   "testdata",
		dbname: "db_bar",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		dir := tt.path
		if dir == "" {
			dir = tempDir(t)
			defer rmdir(t, dir)
		}
		fs := tt.fs
		if fs == nil {
			fs = filesystem.Default()
		}
		c := &client{root: dir, fs: fs}
		db, err := c.newDB(tt.dbname)
		if err != nil {
			t.Fatal(err)
		}
		sec, err := db.Security(context.Background())
		testy.StatusErrorRE(t, tt.err, tt.status, err)
		if d := testy.DiffAsJSON(testy.Snapshot(t), sec); d != nil {
			t.Error(d)
		}
	})
}
