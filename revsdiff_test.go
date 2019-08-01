package fs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"
	"github.com/go-kivik/kivik/driver"
)

func TestRevsDiff(t *testing.T) {
	type tt struct {
		path, dbname string
		revMap       interface{}
		status       int
		err          string
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
		dbname: "db.foo",
		revMap: map[string][]string{
			"yamltest": {"3-", "2-xxx", "1-oink"},
			"autorev":  {"6-", "5-", "4-"},
			"newdoc":   {"1-asdf"},
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
			dbName: tt.dbname,
		}
		rows, err := db.RevsDiff(context.Background(), tt.revMap)
		testy.StatusErrorRE(t, tt.err, tt.status, err)
		result := make(map[string]json.RawMessage)
		var row driver.Row
		for {
			err := rows.Next(&row)
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			result[row.ID] = row.Value
		}
		if d := diff.AsJSON(&diff.File{Path: "testdata/" + testy.Stub(t)}, result); d != nil {
			t.Error(d)
		}
	})
}
