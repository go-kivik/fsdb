package fs

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/go-kivik/kivik/driver"
)

type changes struct {
	db    *db
	ctx   context.Context
	infos []os.FileInfo
}

var _ driver.Changes = &changes{}

func (c *changes) ETag() string    { return "" }
func (c *changes) LastSeq() string { return "" }
func (c *changes) Pending() int64  { return 0 }

func (c *changes) Next(ch *driver.Change) error {
	for {
		if len(c.infos) == 0 {
			return io.EOF
		}
		candidate := c.infos[len(c.infos)-1]
		c.infos = c.infos[:len(c.infos)-1]
		if candidate.IsDir() {
			continue
		}
		if strings.HasSuffix(candidate.Name(), ".json") {
			docid := strings.TrimSuffix(candidate.Name(), ".json")
			rev, err := c.db.currentRev(candidate.Name())
			if err != nil {
				return err
			}
			ch.ID = docid
			ch.Changes = []string{rev}
			return nil
		}
	}
}

func (c *changes) Close() error {
	return nil
}

func (d *db) Changes(ctx context.Context, _ map[string]interface{}) (driver.Changes, error) {
	f, err := os.Open(d.path())
	if err != nil {
		return nil, err
	}
	dir, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}
	return &changes{
		db:    d,
		ctx:   ctx,
		infos: dir,
	}, nil
}
