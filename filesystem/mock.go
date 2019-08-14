package filesystem

import "os"

// MockFS allows mocking a filesystem.
type MockFS struct {
	MkdirFunc    func(string, os.FileMode) error
	MkdirAllFunc func(string, os.FileMode) error
	CreateFunc   func(string) (File, error)
	OpenFunc     func(string) (File, error)
	StatFunc     func(string) (os.FileInfo, error)
	TempFileFunc func(string, string) (File, error)
	RenameFunc   func(string, string) error
}

var _ Filesystem = &MockFS{}

// Mkdir calls fs.MkdirFunc
func (fs *MockFS) Mkdir(name string, perm os.FileMode) error {
	return fs.MkdirFunc(name, perm)
}

// MkdirAll calls fs.MkdirAllFunc
func (fs *MockFS) MkdirAll(path string, perm os.FileMode) error {
	return fs.MkdirAllFunc(path, perm)
}

// Open calls fs.OpenFunc
func (fs *MockFS) Open(name string) (File, error) {
	return fs.OpenFunc(name)
}

// Create calls fs.CreateFunc
func (fs *MockFS) Create(name string) (File, error) {
	return fs.CreateFunc(name)
}

// Stat calls fs.StatFunc
func (fs *MockFS) Stat(name string) (os.FileInfo, error) {
	return fs.StatFunc(name)
}

// TempFile calls fs.TempFileFunc
func (fs *MockFS) TempFile(dir, pattern string) (File, error) {
	return fs.TempFileFunc(dir, pattern)
}

// Rename calls fs.RenameFunc
func (fs *MockFS) Rename(oldpath, newpath string) error {
	return fs.RenameFunc(oldpath, newpath)
}
