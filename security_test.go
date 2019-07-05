package fs

import (
	"context"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"
)

func TestSecurity(t *testing.T) {
	type tt struct {
		path, dbname string
		status       int
		err          string
	}
	tests := testy.NewTable()
	tests.Add("no security object", tt{})
	tests.Add("json security obj", tt{
		path:   "testdata",
		dbname: "db.foo",
	})
	tests.Add("yaml security obj", tt{
		path:   "testdata",
		dbname: "db.bar",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		dir := tt.path
		if dir == "" {
			dir = tempDir(t)
			defer rmdir(t, dir)
		}
		db := &db{
			client: &client{root: dir},
			dbName: tt.dbname,
		}
		sec, err := db.Security(context.Background())
		testy.StatusErrorRE(t, tt.err, tt.status, err)
		if d := diff.AsJSON(&diff.File{Path: "testdata/" + testy.Stub(t)}, sec); d != nil {
			t.Error(d)
		}
	})
}
