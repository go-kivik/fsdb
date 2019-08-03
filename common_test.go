package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/otiai10/copy"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := ioutil.TempDir("", "kivik-fsdb-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func rmdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
}

// copyDir recursively copies the contents of path to a new temporary dir, whose
// path is returned. The depth argument controls how deeply path is placed into
// the temp dir. Examples:
//
//  copyDir(t, "/foo/bar/baz", 0) // copies /foo/bar/baz/* to /tmp-XXX/*
//  copyDir(t, "/foo/bar/baz", 1) // copies /foo/bar/baz/* to /tmp-XXX/baz/*
//  copyDir(t, "/foo/bar/baz", 3) // copies /foo/bar/baz/* to /tmp-XXX/foo/bar/baz/*
func copyDir(t *testing.T, source string, depth int) string { // nolint: unparam
	t.Helper()
	tmpdir := tempDir(t)
	target := tmpdir
	if depth > 0 {
		parts := strings.Split(source, string(filepath.Separator))
		if len(parts) < depth {
			t.Fatalf("Depth of %d specified, but path only has %d parts", depth, len(parts))
		}
		target = filepath.Join(append([]string{tmpdir}, parts[len(parts)-depth:]...)...)
		if err := os.MkdirAll(target, 0777); err != nil {
			t.Fatal(err)
		}
	}
	if err := copy.Copy(source, target); err != nil {
		t.Fatal(err)
	}
	return tmpdir
}
