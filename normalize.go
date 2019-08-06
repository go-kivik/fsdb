package fs

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-kivik/fsdb/decoder"
	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/fsdb/internal"
	"github.com/go-kivik/kivik"
	"golang.org/x/xerrors"
)

type normalDoc struct {
	ID          string                 `json:"_id"`
	Rev         internal.Rev           `json:"_rev,omitempty"`
	Attachments internal.Attachments   `json:"_attachments,omitempty"`
	RevsInfo    []internal.RevsInfo    `json:"_revs_info,omitempty"`
	Revisions   *internal.Revisions    `json:"_revisions,omitempty"`
	Data        map[string]interface{} `json:"-"`
	Path        string                 `json:"-"`
}

func (d *normalDoc) MarshalJSON() ([]byte, error) {
	for key := range d.Data {
		if key[0] == '_' {
			return nil, xerrors.Errorf("Bad special document member: %s", key)
		}
	}
	var data []byte
	if len(d.Data) > 0 {
		var err error
		if data, err = json.Marshal(d.Data); err != nil {
			return nil, err
		}
	}
	doc, err := json.Marshal(*d)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		doc[len(doc)-1] = ','
		return append(doc, data[1:]...), nil
	}
	return doc, nil
}

var reservedKeys = map[string]struct{}{
	"_id":          {},
	"_rev":         {},
	"_attachments": {},
	"_revisions":   {},
	"_revs_info":   {},
}

func (d *normalDoc) UnmarshalJSON(p []byte) error {
	doc := struct {
		ID          string               `json:"_id"`
		Rev         internal.Rev         `json:"_rev,omitempty"`
		Attachments internal.Attachments `json:"_attachments,omitempty"`
		RevsInfo    []internal.RevsInfo  `json:"_revs_info,omitempty"`
		Revisions   *internal.Revisions  `json:"_revisions,omitempty"`
	}{}
	if err := json.Unmarshal(p, &doc); err != nil {
		return err
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(p, &data); err != nil {
		return err
	}
	for key := range data {
		if key[0] == '_' {
			if _, ok := reservedKeys[key]; !ok {
				return xerrors.Errorf("Bad special document member: %s", key)
			}
			delete(data, key)
		}
	}
	d.ID = doc.ID
	d.Rev = doc.Rev
	d.Attachments = doc.Attachments
	d.Revisions = doc.Revisions
	d.RevsInfo = doc.RevsInfo
	d.Data = data
	return nil
}

func (d *normalDoc) cleanup() error {
	var err error
	for _, a := range d.Attachments {
		if e := a.Cleanup(); e != nil {
			err = e
		}
	}
	return err
}

func normalizeDoc(i interface{}) (*normalDoc, error) {
	data, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	doc := &normalDoc{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, &kivik.Error{HTTPStatus: http.StatusBadRequest, Err: err}
	}
	return doc, nil
}

func (d *normalDoc) revsInfo() []internal.RevsInfo {
	// This driver only ever stores the current leaf, so we only ever
	// return one revision here: the current one.
	return []internal.RevsInfo{
		{
			Rev:    d.Rev.String(),
			Status: "available",
		},
	}
}

func (d *normalDoc) revisions() *internal.Revisions {
	if d.Revisions != nil {
		return d.Revisions
	}
	histSize := d.Rev.Seq
	if histSize > revsLimit {
		histSize = revsLimit
	}
	var ids []string
	if d.Rev.Sum == "" {
		ids = make([]string, int(histSize))
	} else {
		ids = []string{d.Rev.Sum}
	}
	return &internal.Revisions{
		Start: d.Rev.Seq,
		IDs:   ids,
	}
}

func readDoc(path string) (*normalDoc, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, kerr(err)
	}
	ext := filepath.Ext(path)[1:]
	i, err := decoder.Decode(f, ext)
	if err != nil {
		return nil, err
	}
	doc, err := normalizeDoc(i)
	if err != nil {
		return nil, err
	}
	doc.Path = path
	return doc, nil
}

func (d *db) openDoc(docID, rev string) (filesystem.File, string, error) {
	base := id2basename(docID)
	for _, ext := range decoder.Extensions() {
		filename := base + "." + ext
		if rev != "" {
			currev, err := d.currentRev(filename, ext)
			if err != nil && !os.IsNotExist(err) {
				return nil, "", err
			}
			if currev != rev {
				revFilename := "." + base + "/" + rev + "." + ext
				f, err := d.fs.Open(d.path(revFilename))
				if !os.IsNotExist(err) {
					return f, ext, err
				}
				continue
			}
		}
		f, err := d.fs.Open(d.path(filename))
		if !os.IsNotExist(err) {
			return f, ext, kerr(err)
		}
	}
	if rev == "" {
		return breakRevTie(d.path("." + base))
	}
	return nil, "", errNotFound
}

type tiedRev struct {
	*internal.Rev
	path string
	ext  string
}

func breakRevTie(path string) (*os.File, string, error) {
	dir, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = errNotFound
		}
		return nil, "", kerr(err)
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, "", kerr(err)
	}
	revs := make([]*tiedRev, 0, len(files))
	for _, info := range files {
		if info.IsDir() {
			continue
		}
		ext := filepath.Ext(info.Name())
		base := strings.TrimSuffix(info.Name(), ext)
		rev := new(internal.Rev)
		if err := rev.UnmarshalText([]byte(base)); err != nil {
			// Ignore unrecognized files
			continue
		}
		revs = append(revs, &tiedRev{
			Rev:  rev,
			path: filepath.Join(path, info.Name()),
			ext:  ext[1:],
		})
	}

	if len(revs) == 0 {
		return nil, "", errNotFound
	}

	sort.Slice(revs, func(i, j int) bool {
		return revs[i].Seq > revs[j].Seq ||
			(revs[i].Seq == revs[j].Seq && revs[i].Sum > revs[j].Sum)
	})

	winner := revs[0]

	f, err := os.Open(winner.path)
	return f, winner.ext, kerr(err)
}

func (d *db) readDoc(docID, rev string) (*normalDoc, error) {
	f, ext, err := d.openDoc(docID, rev)
	if err != nil {
		return nil, err
	}
	defer f.Close() // nolint: errcheck
	i, err := decoder.Decode(f, ext)
	if err != nil {
		return nil, err
	}
	doc, err := normalizeDoc(i)
	if err != nil {
		return nil, err
	}
	doc.Path = f.Name()
	return doc, nil
}

func (d *db) get(_ context.Context, docID string, opts map[string]interface{}) (*normalDoc, error) {
	rev, _ := opts["rev"].(string)
	ndoc, err := d.readDoc(docID, rev)
	if err != nil {
		return nil, kerr(err)
	}
	if ndoc.Rev.IsZero() {
		ndoc.Rev.Increment()
	}
	ndoc.Revisions = ndoc.revisions()
	return ndoc, nil
}
