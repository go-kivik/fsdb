package yamldec

import (
	"io"

	"github.com/go-kivik/fsdb/decoder"
	"github.com/go-kivik/fsdb/internal"
	"github.com/icza/dyno"
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
	return dyno.ConvertMapI2MapS(doc).(map[string]interface{}), err
}

func (d *dec) Rev(r io.Reader) (internal.Rev, error) {
	doc := internal.RevDoc{}
	err := yaml.NewDecoder(r).Decode(&doc)
	return doc.Rev, err
}
