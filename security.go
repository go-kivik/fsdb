/*
Security Objects

This driver supports fetching and storing security objects, but completely
ignores them for access control. This support is intended only for the purpose
of syncing to/from CouchDB instances.
*/

package fs

import (
	"context"

	"github.com/go-kivik/kivik/v3/driver"
)

func (d *db) Security(ctx context.Context) (*driver.Security, error) {
	return d.cdb.ReadSecurity(ctx, d.path())
}

func (d *db) SetSecurity(_ context.Context, _ *driver.Security) error {
	// FIXME: Unimplemented
	return notYetImplemented
}
