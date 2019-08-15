package fs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/kivik/driver"
)

func TestRevsDiff(t *testing.T) {
	type tt struct {
		ctx          context.Context
		fs           filesystem.Filesystem
		path, dbname string
		revMap       interface{}
		status       int
		err          string
		rowStatus    int
		rowErr       string
	}
	tests := testy.NewTable()
	tests.Add("invalid revMap", tt{
		revMap: make(chan int),
		status: http.StatusBadRequest,
		err:    "json: unsupported type: chan int",
	})
	tests.Add("empty map", tt{
		revMap: map[string][]string{},
	})
	tests.Add("real test", tt{
		path:   "testdata",
		dbname: "db_foo",
		revMap: map[string][]string{
			"yamltest": {"3-", "2-xxx", "1-oink"},
			"autorev":  {"6-", "5-", "4-"},
			"newdoc":   {"1-asdf"},
		},
	})
	tests.Add("cancelled context", func(t *testing.T) interface{} {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return tt{
			ctx:    ctx,
			path:   "testdata",
			dbname: "db_foo",
			revMap: map[string][]string{
				"yamltest": {"3-", "2-xxx", "1-oink"},
				"autorev":  {"6-", "5-", "4-"},
				"newdoc":   {"1-asdf"},
			},
			rowStatus: http.StatusInternalServerError,
			rowErr:    "context canceled",
		}
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
		db := c.newDB(tt.dbname)
		ctx := tt.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		rows, err := db.RevsDiff(ctx, tt.revMap)
		testy.StatusErrorRE(t, tt.err, tt.status, err)
		result := make(map[string]json.RawMessage)
		var row driver.Row
		var rowErr error
		for {
			err := rows.Next(&row)
			if err == io.EOF {
				break
			}
			if err != nil {
				rowErr = err
				break
			}
			result[row.ID] = row.Value
		}
		testy.StatusErrorRE(t, tt.rowErr, tt.rowStatus, rowErr)
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}
