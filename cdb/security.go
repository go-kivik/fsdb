package cdb

import (
	"context"
	"os"
	"path/filepath"

	"github.com/go-kivik/fsdb/cdb/decode"
	"github.com/go-kivik/kivik/driver"
)

// ReadSecurity reads the _security.{ext} document from path.
func (fs *FS) ReadSecurity(ctx context.Context, path string) (*driver.Security, error) {
	sec := new(driver.Security)
	for _, ext := range decode.Extensions() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		f, err := fs.fs.Open(filepath.Join(path, "_security."+ext))
		if err == nil {
			defer f.Close() // nolint: errcheck
			err := decode.Decode(f, ext, sec)
			return sec, err
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return sec, nil
}
