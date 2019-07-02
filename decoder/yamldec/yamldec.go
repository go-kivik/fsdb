package yamldec

import (
	"io"

	"github.com/go-kivik/fsdb/decoder"
	"gopkg.in/yaml.v2"
)

type dec struct{}

func init() {
	decoder.Register(&dec{})
}

func (d *dec) Extensions() []string {
	return []string{"yaml", "yml"}
}

func (d *dec) Decode(r io.Reader) (map[string]interface{}, error) {
	doc := map[string]interface{}{}
	err := yaml.NewDecoder(r).Decode(&doc)
	return doc, err
}
