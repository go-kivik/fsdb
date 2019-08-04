package fs

import (
	"net/http"
	"testing"

	"github.com/go-kivik/kivik"
)

func TestCreateDBUnauthorized(t *testing.T) {
	path := "/this/better/not/exist"
	_, err := kivik.New("fs", path)
	if err == nil {
		t.Errorf("Expected error attempting to create FS database in '%s'\n", path)
		return
	}
	if kivik.StatusCode(err) != http.StatusUnauthorized {
		t.Errorf("Expected Unauthorized error trying to create FS database in '%s', but got %s\n", path, err)
	}
}
