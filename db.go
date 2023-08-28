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
	"net/http"
	"path/filepath"

	"github.com/go-kivik/fsdb/v4/cdb"
	"github.com/go-kivik/fsdb/v4/filesystem"
	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

type db struct {
	*client
	dbPath, dbName string
	fs             filesystem.Filesystem
	cdb            *cdb.FS
}

var _ driver.DB = &db{}

var notYetImplemented = &kivik.Error{Status: http.StatusNotImplemented, Message: "kivik: not yet implemented in fs driver"}

func (d *db) path(parts ...string) string {
	return filepath.Join(append([]string{d.dbPath}, parts...)...)
}

func (d *db) AllDocs(_ context.Context, _ map[string]interface{}) (driver.Rows, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) Query(ctx context.Context, ddoc, view string, opts map[string]interface{}) (driver.Rows, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) CreateDoc(_ context.Context, doc interface{}, opts map[string]interface{}) (docID, rev string, err error) {
	// FIXME: Unimplemented
	return "", "", notYetImplemented
}

func (d *db) Delete(_ context.Context, docID string, opts map[string]interface{}) (newRev string, err error) {
	// FIXME: Unimplemented
	return "", notYetImplemented
}

func (d *db) Stats(_ context.Context) (*driver.DBStats, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) CompactView(_ context.Context, _ string) error {
	// FIXME: Unimplemented
	return notYetImplemented
}

func (d *db) ViewCleanup(_ context.Context) error {
	// FIXME: Unimplemented
	return notYetImplemented
}

func (d *db) BulkDocs(_ context.Context, _ []interface{}) ([]driver.BulkResult, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) PutAttachment(_ context.Context, _ string, _ *driver.Attachment, _ map[string]interface{}) (string, error) {
	// FIXME: Unimplemented
	return "", notYetImplemented
}

func (d *db) GetAttachment(ctx context.Context, docID, filename string, opts map[string]interface{}) (*driver.Attachment, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) DeleteAttachment(ctx context.Context, docID, filename string, opts map[string]interface{}) (newRev string, err error) {
	// FIXME: Unimplemented
	return "", notYetImplemented
}
