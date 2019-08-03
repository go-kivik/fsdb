package fs

import (
	"context"
	"os"
)

func (d *db) Compact(ctx context.Context) error {
	dir, err := os.Open(d.path())
	if err != nil {
		return kerr(err)
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	for _, i := range files {
		if err := os.RemoveAll(d.path(i.Name())); err != nil {
			return err
		}
	}

	return nil
}
