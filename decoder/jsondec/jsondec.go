package jsondec

import (
	"encoding/json"
	"io"

	"github.com/go-kivik/fsdb/decoder"
	"github.com/go-kivik/fsdb/internal"
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

func (d *dec) Rev(r io.Reader) (internal.Rev, error) {
	doc := internal.RevDoc{}
	err := json.NewDecoder(r).Decode(&doc)
	return doc.Rev, err
}
