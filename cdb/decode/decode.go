package decode

import (
	"io"
	"os"
	"strings"

	"github.com/go-kivik/fsdb/filesystem"
	"golang.org/x/xerrors"
)

type decoder interface {
	Decode(io.Reader, interface{}) error
}

var decoders = map[string]decoder{
	"json": &jsonDecoder{},
	"yaml": &yamlDecoder{},
	"yml":  &yamlDecoder{},
}

// OpenAny attempts to open base + any supported extension. It returns the open
// file, the matched extension, or an error.
func OpenAny(fs filesystem.Filesystem, base string) (f filesystem.File, ext string, err error) {
	for ext = range decoders {
		f, err = fs.Open(base + "." + ext)
		if err == nil || !os.IsNotExist(err) {
			return
		}
	}
	return
}

// Decode decodes r according to ext's registered decoder, into i.
func Decode(r io.Reader, ext string, i interface{}) error {
	ext = strings.TrimPrefix(ext, ".")
	dec, ok := decoders[ext]
	if !ok {
		return xerrors.Errorf("No decoder for %s", ext)
	}
	return dec.Decode(r, i)
}
