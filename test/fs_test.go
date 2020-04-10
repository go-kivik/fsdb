// +build !js

package test

import (
	"io/ioutil"
	"os"
	"testing"

	_ "github.com/go-kivik/fsdb/v4"
	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4"
	"github.com/go-kivik/kiviktest/v4/kt"
)

func init() {
	RegisterFSSuite()
}

func TestFS(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kivik.test.")
	if err != nil {
		t.Errorf("Failed to create temp dir to test FS driver: %s\n", err)
		return
	}
	defer os.RemoveAll(tempDir) // To clean up after tests
	client, err := kivik.New("fs", tempDir)
	if err != nil {
		t.Errorf("Failed to connect to FS driver: %s\n", err)
		return
	}
	clients := &kt.Context{
		RW:    true,
		Admin: client,
	}
	kiviktest.RunTestsInternal(clients, kiviktest.SuiteKivikFS, t)
}
