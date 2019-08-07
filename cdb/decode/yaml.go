package decode

import (
	"io"

	yaml "gopkg.in/yaml.v2"
)

type yamlDecoder struct{}

func (d yamlDecoder) Decode(r io.Reader, i interface{}) error {
	return yaml.NewDecoder(r).Decode(i)
}
