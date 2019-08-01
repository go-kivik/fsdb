package fs

import (
	"net/http"
	"os"

	"github.com/go-kivik/kivik"
)

func kerr(err error) error {
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		return &kivik.Error{HTTPStatus: http.StatusNotFound, Err: err}
	}
	if os.IsPermission(err) {
		return &kivik.Error{HTTPStatus: http.StatusForbidden, Err: err}
	}
	return err
}
