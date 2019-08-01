package fs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-kivik/kivik/driver"
	"gitlab.com/flimzy/ale/httperr"
)

var _ driver.RevsDiffer = &db{}

func toRevmap(i interface{}) (map[string][]string, error) {
	switch t := i.(type) {
	case map[string][]string:
		return t, nil
	}
	encoded, err := json.Marshal(i)
	if err != nil {
		return nil, httperr.WithStatus(http.StatusBadRequest, err)
	}
	revmap := make(map[string][]string)
	err = json.Unmarshal(encoded, &revmap)
	return revmap, httperr.WithStatus(http.StatusBadRequest, err)
}

func (d *db) RevsDiff(ctx context.Context, revMap interface{}) (driver.Rows, error) {
	_, err := toRevmap(revMap)
	if err != nil {
		return nil, err
	}
	return &revDiffRows{}, nil
}

type revDiffRows struct{}

var _ driver.Rows = &revDiffRows{}

func (r *revDiffRows) Close() error {
	return nil
}

func (r *revDiffRows) Next(row *driver.Row) error {
	return io.EOF
}

func (r *revDiffRows) Offset() int64     { return 0 }
func (r *revDiffRows) TotalRows() int64  { return 0 }
func (r *revDiffRows) UpdateSeq() string { return "" }
