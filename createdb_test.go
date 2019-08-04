package fs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/kivik/driver"
	"gitlab.com/flimzy/testy"
)

func TestCreateDB(t *testing.T) {
	type tt struct {
		driver *fsDriver
		path   string
		status int
		err    string
		want   *client
	}
	tests := testy.NewTable()
	tests.Add("not a dir", tt{
		driver: &fsDriver{},
		path:   "testdata/foo.txt",
		status: http.StatusBadRequest,
		err:    "testdata/foo.txt is not a directory",
	})
	tests.Add("not exist", tt{
		driver: &fsDriver{},
		path:   "testdata/doesnotexist",
		status: http.StatusNotFound,
		err:    "stat testdata/doesnotexist: no such file or directory",
	})
	tests.Add("success", tt{
		driver: &fsDriver{},
		path:   "testdata",
		want: &client{
			version: &driver.Version{
				Version:     Version,
				Vendor:      Vendor,
				RawResponse: json.RawMessage(fmt.Sprintf(`{"version":"%s","vendor":{"name":"%s"}}`, Version, Vendor)),
			},
			root: "testdata",
			fs:   filesystem.Default(),
		},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		client, err := tt.driver.NewClient(tt.path)
		testy.StatusError(t, tt.err, tt.status, err)
		if d := testy.DiffInterface(tt.want, client); d != nil {
			t.Error(d)
		}
	})
}
