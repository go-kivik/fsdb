package cdb

import (
	"net/http"
	"testing"

	"github.com/go-kivik/fsdb/v4/filesystem"
	"github.com/go-kivik/kivik/v4"
	"gitlab.com/flimzy/testy"
)

func TestAttachmentsIterator(t *testing.T) {
	type tt struct {
		r      *Revision
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("missing attachment", tt{
		r: &Revision{
			options: kivik.Options{
				"attachments": true,
			},
			RevMeta: RevMeta{
				Attachments: map[string]*Attachment{
					"notfound.txt": {
						fs:   filesystem.Default(),
						path: "/somewhere/notfound.txt",
					},
				},
			},
		},
		status: http.StatusInternalServerError,
		err:    "open /somewhere/notfound.txt: no such file or directory",
	})
	tests.Run(t, func(t *testing.T, tt tt) {
		_, err := tt.r.AttachmentsIterator()
		testy.StatusError(t, tt.err, tt.status, err)
	})
}
