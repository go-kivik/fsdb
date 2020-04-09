package fs

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-kivik/fsdb/v3/filesystem"
	"github.com/go-kivik/kivik/v3/driver"
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
