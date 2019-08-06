package decoder

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"

	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/fsdb/internal"
	"github.com/go-kivik/kivik"
	"github.com/go-kivik/kivik/driver"
)

// Decoder represents an abstract document decoder.
type Decoder interface {
	Extensions() []string
	Decode(io.Reader) (map[string]interface{}, error)
	DecodeSecurity(io.Reader) (*driver.Security, error)
	Rev(io.Reader) (internal.Rev, error)
	DocMeta(io.Reader) (*internal.DocMeta, error)
}

var decoders = map[string]Decoder{}
var extensions = []string{}

// Register registers the specified decoder for the extensions it supports.
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

// Decode passes r to the appropriate decoder for ext, to decode the document.
func Decode(r io.Reader, ext string) (map[string]interface{}, error) {
	dec, ok := decoders[ext]
	if !ok {
		return nil, fmt.Errorf("No decoder for ext '%s'", ext)
	}
	return dec.Decode(r)
}

// DecodeSecurity passes r to the appropriate decoder for ext, to decode the
// security document.
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

// DocMeta extracts the document metadata from r, based on the decoder registered
// for ext.
func DocMeta(r io.Reader, ext string) (*internal.DocMeta, error) {
	dec, ok := decoders[ext]
	if !ok {
		return nil, fmt.Errorf("No decoder for ext '%s'", ext)
	}
	return dec.DocMeta(r)
}

// ReadRev cycles through all registered extensions, and if found, reads the
// rev and returns it.
func ReadRev(fs filesystem.Filesystem, base string) (internal.Rev, error) {
	for _, ext := range Extensions() {
		f, err := fs.Open(base + "." + ext)
		switch {
		case err == nil:
			return Rev(f, ext)
		case !os.IsNotExist(err):
			return internal.Rev{}, err
		}
	}
	return internal.Rev{}, &kivik.Error{HTTPStatus: http.StatusNotFound, Message: "missing"}
}

// Extensions returns the registered file extensions.
func Extensions() []string {
	return extensions
}

// ReadDocMeta reads document metadata from the document. extHint, if provided,
// short-circuits the extension detection.
func ReadDocMeta(fs filesystem.Filesystem, base string, extHint ...string) (*internal.DocMeta, error) {
	if len(extHint) > 0 {
		ext := extHint[0]
		f, err := fs.Open(base + "." + ext)
		if err != nil {
			return nil, err
		}
		return DocMeta(f, ext)
	}
	for _, ext := range Extensions() {
		f, err := fs.Open(base + "." + ext)
		switch {
		case err == nil:
			return DocMeta(f, ext)
		case !os.IsNotExist(err):
			return nil, err
		}
	}
	return nil, &kivik.Error{HTTPStatus: http.StatusNotFound, Message: "missing"}
}
