// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package fs

import (
	"context"
	"io"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
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
			dbName: "db_foo",
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
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}
