package cdb

import (
	"errors"
	"net/http"
	"os"

	"github.com/go-kivik/kivik"
	"golang.org/x/xerrors"
)

var (
	errNotFound         = &kivik.Error{HTTPStatus: http.StatusNotFound, Message: "missing"}
	errUnrecognizedFile = errors.New("unrecognized file")
	errConflict         = &kivik.Error{HTTPStatus: http.StatusConflict, Message: "Document update conflict."}
)

// missing transforms a NotExist error into a standard CouchDBesque 'missing'
// error. All other values are passed through unaltered.
func missing(err error) error {
	if !os.IsNotExist(err) {
		return err
	}
	return errNotFound
}

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
