package fs

import (
	"net/http"
	"os"

	"github.com/go-kivik/kivik"
	"golang.org/x/xerrors"
)

var (
	errNotFound = &kivik.Error{HTTPStatus: http.StatusNotFound, Message: "missing"}
)

func kerr(err error) error {
	if err == nil {
		return nil
	}
	if xerrors.Is(err, &kivik.Error{}) {
		// Error has already been converted
		return err
	}
	if os.IsNotExist(err) {
		return &kivik.Error{HTTPStatus: http.StatusNotFound, Err: err}
	}
	if os.IsPermission(err) {
		return &kivik.Error{HTTPStatus: http.StatusForbidden, Err: err}
	}
	return err
}
