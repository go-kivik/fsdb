package filesystem

// MockFS allows mocking a filesystem.
type MockFS struct {
	CreateFunc func(string) (File, error)
	OpenFunc   func(string) (File, error)
}

var _ Filesystem = &MockFS{}

// Open calls fs.OpenFunc
func (fs *MockFS) Open(name string) (File, error) {
	return fs.OpenFunc(name)
}

// Create calls fs.CreateFunc
func (fs *MockFS) Create(name string) (File, error) {
	return fs.CreateFunc(name)
}
