package fs

import (
	"context"
	"io"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"

	"github.com/go-kivik/kivik/driver"
)

func TestChanges(t *testing.T) {
	type tt struct {
		db      *db
		options map[string]interface{}
		status  int
		err     string
	}
	tests := testy.NewTable()
	tests.Add("success", tt{
		db: &db{
			client: &client{root: "testdata"},
			dbName: "db.foo",
		},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		changes, err := tt.db.Changes(context.TODO(), tt.options)
		testy.StatusError(t, tt.err, tt.status, err)
		defer changes.Close() // nolint: errcheck
		result := make(map[string][]string)
		ch := &driver.Change{}
		for {
			if err := changes.Next(ch); err != nil {
				if err == io.EOF {
					break
				}
				t.Fatal(err)
			}
			result[ch.ID] = ch.Changes
		}
		if d := diff.AsJSON(&diff.File{Path: "testdata/" + testy.Stub(t)}, result); d != nil {
			t.Error(d)
		}
	})
}
