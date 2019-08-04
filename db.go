package fs

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/kivik/driver"
	"github.com/go-kivik/kivik/errors"
)

type db struct {
	*client
	dbName string
	fs     filesystem.Filesystem
}

var _ driver.DB = &db{}

var notYetImplemented = errors.Status(http.StatusNotImplemented, "kivik: not yet implemented in fs driver")

func (d *db) path(parts ...string) string {
	return filepath.Join(append([]string{d.client.root, d.dbName}, parts...)...)
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

func (d *db) Delete(_ context.Context, docID, rev string, opts map[string]interface{}) (newRev string, err error) {
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

func (d *db) BulkDocs(_ context.Context, _ []interface{}) (driver.BulkResults, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) PutAttachment(_ context.Context, _, _ string, _ *driver.Attachment, _ map[string]interface{}) (string, error) {
	// FIXME: Unimplemented
	return "", notYetImplemented
}

func (d *db) GetAttachment(ctx context.Context, docID, filename string, opts map[string]interface{}) (*driver.Attachment, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) DeleteAttachment(ctx context.Context, docID, rev, filename string, opts map[string]interface{}) (newRev string, err error) {
	// FIXME: Unimplemented
	return "", notYetImplemented
}
