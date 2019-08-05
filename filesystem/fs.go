// Package filesystem provides an abstraction around a filesystem
package filesystem

import (
	"io"
	"os"
)

// Filesystem is a filesystem implemenatation.
type Filesystem interface {
	Mkdir(name string, perm os.FileMode) error
	Open(string) (File, error)
	Create(string) (File, error)
	Stat(string) (os.FileInfo, error)
}

type defaultFS struct{}

var _ Filesystem = &defaultFS{}

func (fs *defaultFS) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (fs *defaultFS) Open(name string) (File, error) {
	return os.Open(name)
}

func (fs *defaultFS) Create(name string) (File, error) {
	return os.Create(name)
}

func (fs *defaultFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// File represents a file object.
type File interface {
	io.Reader
	io.Closer
	io.Writer
	io.ReaderAt
	io.Seeker
	Name() string
	Readdir(int) ([]os.FileInfo, error)
	Stat() (os.FileInfo, error)
}

type defaultFile struct {
	*os.File
}

var _ File = &defaultFile{}

// Default returns the default filesystem implementation.
func Default() Filesystem {
	return &defaultFS{}
}
