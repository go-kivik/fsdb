package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/go-kivik/fsdb/cdb"
	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/kivik"
	"github.com/go-kivik/kivik/driver"
)

const dirMode = os.FileMode(0700)

type fsDriver struct {
	fs filesystem.Filesystem
}

var _ driver.Driver = &fsDriver{}

// Identifying constants
const (
	Version = "0.0.1"
	Vendor  = "Kivik File System Adaptor"
)

func init() {
	kivik.Register("fs", &fsDriver{})
}

type client struct {
	version *driver.Version
	root    string
	fs      filesystem.Filesystem
}

var _ driver.Client = &client{}

func (d *fsDriver) NewClient(dir string) (driver.Client, error) {
	if err := validateRootDir(dir); err != nil {
		return nil, kerr(err)
	}
	fs := d.fs
	if fs == nil {
		fs = filesystem.Default()
	}
	return &client{
		version: &driver.Version{
			Version:     Version,
			Vendor:      Vendor,
			RawResponse: json.RawMessage(fmt.Sprintf(`{"version":"%s","vendor":{"name":"%s"}}`, Version, Vendor)),
		},
		fs:   fs,
		root: dir,
	}, nil
}

func validateRootDir(dir string) error {
	// See if the target path exists, and is a directory
	info, err := os.Stat(dir)
	if err != nil {
		return kerr(err)
	}
	if !info.IsDir() {
		return &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: fmt.Sprintf("%s is not a directory", dir)}
	}
	return nil
}

// Version returns the configured server info.
func (c *client) Version(_ context.Context) (*driver.Version, error) {
	return c.version, nil
}

// Taken verbatim from http://docs.couchdb.org/en/2.0.0/api/database/common.html
var validDBNameRE = regexp.MustCompile("^[a-z_][a-z0-9_$()+/-]*$")

// AllDBs returns a list of all DBs present in the configured root dir.
func (c *client) AllDBs(_ context.Context, _ map[string]interface{}) ([]string, error) {
	files, err := ioutil.ReadDir(c.root)
	if err != nil {
		return nil, err
	}
	filenames := make([]string, 0, len(files))
	for _, file := range files {
		if !validDBNameRE.MatchString(file.Name()) {
			// FIXME #64: Add option to warn about non-matching files?
			continue
		}
		filenames = append(filenames, file.Name())
	}
	return filenames, nil
}

// CreateDB creates a database
func (c *client) CreateDB(ctx context.Context, dbName string, options map[string]interface{}) error {
	exists, err := c.DBExists(ctx, dbName, options)
	if err != nil {
		return err
	}
	if exists {
		return &kivik.Error{HTTPStatus: http.StatusPreconditionFailed, Message: "database already exists"}
	}
	if err := os.Mkdir(c.root+"/"+dbName, dirMode); err != nil {
		return err
	}
	return nil
}

// DBExistsreturns true if the database exists.
func (c *client) DBExists(_ context.Context, dbName string, _ map[string]interface{}) (bool, error) {
	_, err := os.Stat(c.root + "/" + dbName)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// DestroyDB destroys the database
func (c *client) DestroyDB(ctx context.Context, dbName string, options map[string]interface{}) error {
	exists, err := c.DBExists(ctx, dbName, options)
	if err != nil {
		return err
	}
	if !exists {
		return &kivik.Error{HTTPStatus: http.StatusNotFound, Message: "database does not exist"}
	}
	return os.RemoveAll(c.root + "/" + dbName)
}

func (c *client) DB(_ context.Context, dbName string, _ map[string]interface{}) (driver.DB, error) {
	return c.newDB(dbName), nil
}

func (c *client) newDB(dbName string) *db {
	return &db{
		client: c,
		dbName: dbName,
		fs:     c.fs,
		cdb:    cdb.New(filepath.Join(c.root, dbName), c.fs),
	}
}
