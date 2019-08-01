package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-kivik/kivik"
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
	revmap, err := toRevmap(revMap)
	if err != nil {
		return nil, err
	}
	return &revDiffRows{
		ctx:    ctx,
		db:     d,
		revmap: revmap,
	}, nil
}

type revDiffRows struct {
	ctx    context.Context
	db     *db
	revmap map[string][]string
}

var _ driver.Rows = &revDiffRows{}

func (r *revDiffRows) Close() error {
	return nil
}

// maxRev returns the highest key from the map
func maxRev(revs map[string]struct{}) string {
	var max string
	for k := range revs {
		if k > max {
			max = k
		}
	}
	return max
}

func (r *revDiffRows) next() (docID string, missing []string, err error) {
	if len(r.revmap) == 0 {
		return "", nil, io.EOF
	}
	if err := r.ctx.Err(); err != nil {
		return "", nil, err
	}
	revs := map[string]struct{}{}
	for k, v := range r.revmap {
		docID = k
		for _, rev := range v {
			revs[rev] = struct{}{}
		}
		break
	}
	delete(r.revmap, docID)
	for len(revs) > 0 {
		rev := maxRev(revs)
		delete(revs, rev)
		ndoc, err := r.db.get(r.ctx, docID, kivik.Options{
			"rev": rev,
		})
		if kivik.StatusCode(err) == http.StatusNotFound {
			missing = append(missing, rev)
			continue
		}
		if err != nil {
			return "", nil, err
		}
		revisions := ndoc.revisions()
		for i, id := range revisions.IDs {
			rev := fmt.Sprintf("%d-%s", revisions.Start+1-int64(i), id)
			delete(revs, rev)
		}
	}
	if len(missing) == 0 {
		return r.next()
	}
	return docID, missing, nil
}

func (r *revDiffRows) Next(row *driver.Row) error {
	docID, missing, err := r.next()
	if err != nil {
		return err
	}
	row.ID = docID
	row.Value, err = json.Marshal(map[string][]string{
		"missing": missing,
	})
	return err
}

func (r *revDiffRows) Offset() int64     { return 0 }
func (r *revDiffRows) TotalRows() int64  { return 0 }
func (r *revDiffRows) UpdateSeq() string { return "" }
