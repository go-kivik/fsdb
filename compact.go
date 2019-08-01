package fs

import (
	"context"
	"fmt"
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
		fmt.Printf("file: %s\n", i.Name())
	}
	return nil
}
