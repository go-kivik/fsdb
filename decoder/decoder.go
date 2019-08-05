package decoder

import (
	"fmt"
	"io"
	"sort"

	"github.com/go-kivik/fsdb/internal"
	"github.com/go-kivik/kivik/driver"
)

type Decoder interface {
	Extensions() []string
	Decode(io.Reader) (map[string]interface{}, error)
	DecodeSecurity(io.Reader) (*driver.Security, error)
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

func DecodeSecurity(r io.Reader, ext string) (*driver.Security, error) {
	dec, ok := decoders[ext]
	if !ok {
		return nil, fmt.Errorf("No decoder for ext '%s'", ext)
	}
	return dec.DecodeSecurity(r)
}

// Rev extracts the revision from r, based on the decoder registered for ext.
func Rev(r io.Reader, ext string) (internal.Rev, error) {
	dec, ok := decoders[ext]
	if !ok {
		return internal.Rev{}, fmt.Errorf("No decoder for ext '%s'", ext)
	}
	return dec.Rev(r)
}

// Extensions returns the registered file extensions.
func Extensions() []string {
	return extensions
}
