package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/pkg/errors"

	"github.com/go-kivik/kivik"
	"github.com/go-kivik/kivik/driver"
)

const dirMode = os.FileMode(0700)

type fsDriver struct{}

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
}

var _ driver.Client = &client{}

func (d *fsDriver) NewClient(dir string) (driver.Client, error) {
	if err := validateRootDir(dir); err != nil {
		if os.IsPermission(errors.Cause(err)) {
			return nil, &kivik.Error{HTTPStatus: http.StatusUnauthorized, Message: "access denied"}
		}
		return nil, err
	}
	return &client{
		version: &driver.Version{
			Version:     Version,
			Vendor:      Vendor,
			RawResponse: json.RawMessage(fmt.Sprintf(`{"version":"%s","vendor":{"name":"%s"}}`, Version, Vendor)),
		},
		root: dir,
	}, nil
}

func validateRootDir(dir string) error {
	// See if the target path exists, and is a directory
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, dirMode); err != nil {
			return errors.Wrapf(err, "failed to create dir '%s'", dir)
		}
		return nil
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
		if file.Name()[0] == '.' {
			// As a special case, we skip over dot files
			continue
		}
		if !validDBNameRE.MatchString(file.Name()) {
			// Warn about bad filenames
			fmt.Printf("kivik: Filename does not conform to database name standards: %s/%s\n", c.root, file.Name())
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
	if err = os.RemoveAll(c.root + "/" + dbName); err != nil {
		return err
	}
	return nil
}

func (c *client) DB(_ context.Context, dbName string, _ map[string]interface{}) (driver.DB, error) {
	return &db{
		client: c,
		dbName: dbName,
	}, nil
}
