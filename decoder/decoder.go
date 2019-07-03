package decoder

import (
	"fmt"
	"io"
	"sort"

	"github.com/go-kivik/fsdb/internal"
)

type Decoder interface {
	Extensions() []string
	Decode(io.Reader) (map[string]interface{}, error)
	Rev(io.Reader) (internal.Rev, error)
}

var decoders = map[string]Decoder{}
var extensions = []string{}

func Register(dec Decoder) {
	exts := dec.Extensions()
	for _, ext := range exts {
		if _, ok := decoders[ext]; ok {
			panic(fmt.Sprintf("Decoder for extension '%s' already registered", ext))
		}
		decoders[ext] = dec
	}
	extensions = append(extensions, exts...)
	sort.Strings(extensions)
}

func Decode(r io.Reader, ext string) (map[string]interface{}, error) {
	dec, ok := decoders[ext]
	if !ok {
		return nil, fmt.Errorf("No decoder for ext '%s'", ext)
	}
	return dec.Decode(r)
}

func Rev(r io.Reader, ext string) (string, error) {
	dec, ok := decoders[ext]
	if !ok {
		return "", fmt.Errorf("No decoder for ext '%s'", ext)
	}
	rev, err := dec.Rev(r)
	return rev.String(), err
}

func Extensions() []string {
	return extensions
}
