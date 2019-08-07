package jsondec

import (
	"encoding/json"
	"io"

	"github.com/go-kivik/fsdb/decoder"
	"github.com/go-kivik/fsdb/internal"
	"github.com/go-kivik/kivik/driver"
)

type dec struct{}

func init() {
	decoder.Register(&dec{})
}

func (d *dec) Extensions() []string {
	return []string{"json"}
}

func (d *dec) Decode(r io.Reader) (map[string]interface{}, error) {
	doc := map[string]interface{}{}
	err := json.NewDecoder(r).Decode(&doc)
	return doc, err
}

func (d *dec) DecodeSecurity(r io.Reader) (*driver.Security, error) {
	sec := &driver.Security{}
	err := json.NewDecoder(r).Decode(sec)
	return sec, err
}

func (d *dec) Rev(r io.Reader) (internal.RevID, error) {
	meta, err := d.DocMeta(r)
	return meta.Rev, err
}

func (d *dec) DocMeta(r io.Reader) (*internal.DocMeta, error) {
	meta := &internal.DocMeta{}
	err := json.NewDecoder(r).Decode(&meta)
	return meta, err
}
