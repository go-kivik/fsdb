package decode

import (
	"encoding/json"
	"io"
)

type jsonDecoder struct{}

func (d *jsonDecoder) Decode(r io.Reader, i interface{}) error {
	return json.NewDecoder(r).Decode(i)
}
