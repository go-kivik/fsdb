package jsondec

import (
	"encoding/json"
	"io"

	"github.com/go-kivik/fsdb/decoder"
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
