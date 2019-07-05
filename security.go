/*
Security Objects

This driver supports fetching and storing security objects, but completely
ignores them for access control. This support is intended only for the purpose
of syncing to/from CouchDB instances.
*/

package fs

import (
	"context"
	"os"

	"github.com/go-kivik/fsdb/decoder"
	"github.com/go-kivik/kivik/driver"
)

func (d *db) Security(ctx context.Context) (*driver.Security, error) {
	for _, ext := range decoder.Extensions() {
		f, err := os.Open(d.path("_security." + ext))
		if err == nil {
			defer f.Close() // nolint: errcheck
			return decoder.DecodeSecurity(f, ext)
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return &driver.Security{}, nil
}

func (d *db) SetSecurity(_ context.Context, _ *driver.Security) error {
	// FIXME: Unimplemented
	return notYetImplemented
}
