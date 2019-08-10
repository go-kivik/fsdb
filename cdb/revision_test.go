package cdb

import (
	"testing"

	"github.com/go-kivik/kivik"
	"gitlab.com/flimzy/testy"
)

func TestNewRevision(t *testing.T) {
	type tt struct {
		i      interface{}
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("simple", tt{
		i: map[string]interface{}{
			"_rev":  "1-xxx",
			"value": "foo",
		},
	})
	tests.Add("with attachments", tt{
		i: map[string]interface{}{
			"_rev": "3-asdf",
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"content_type": "text/plain",
					"data":         []byte("This is some content"),
				},
			},
		},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		fs := &FS{}
		rev, err := fs.NewRevision(tt.i)
		testy.StatusError(t, tt.err, tt.status, err)
		rev.options = kivik.Options{
			"revs":          true,
			"attachments":   true,
			"header:accept": "application/json",
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), rev); d != nil {
			t.Error(d)
		}
	})
}
