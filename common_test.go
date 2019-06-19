package fs

import (
	"io/ioutil"
	"os"
	"testing"
)

func tempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "kivik-fsdb-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func rmdir(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
}
