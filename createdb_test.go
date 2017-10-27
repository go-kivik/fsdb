package fs

import (
	"context"
	"testing"

	"github.com/flimzy/kivik"
)

func TestCreateDBUnauthorized(t *testing.T) {
	path := "/this/better/not/exist"
	_, err := kivik.New(context.Background(), "fs", path)
	if err == nil {
		t.Errorf("Expected error attempting to create FS database in '%s'\n", path)
		return
	}
	if kivik.StatusCode(err) != kivik.StatusUnauthorized {
		t.Errorf("Expected Unauthorized error trying to create FS database in '%s', but got %s\n", path, err)
	}
}
