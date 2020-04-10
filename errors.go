package fs

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/go-kivik/kivik/v4"
)

func illegalDBName(dbname string) error {
	return &kivik.Error{HTTPStatus: http.StatusBadRequest,
		Message: fmt.Sprintf("Name: '%s'. Only lowercase characters (a-z), digits (0-9), and any of the characters _, $, (, ), +, -, and / are allowed. Must begin with a letter.", dbname)}
}

func kerr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, &kivik.Error{}) {
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
