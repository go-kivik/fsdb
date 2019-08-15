package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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

func parseFileURL(dir string) (string, error) {
	parsed, err := url.Parse(dir)
	if parsed.Scheme != "" && parsed.Scheme != "file" {
		return "", &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: fmt.Sprintf("Unsupported URL scheme '%s'. Wrong driver?", parsed.Scheme)}
	}
	if !strings.HasPrefix("file://", dir) {
		return dir, nil
	}
	if err != nil {
		return "", err
	}
	return parsed.Path, nil
}

func (d *fsDriver) NewClient(dir string) (driver.Client, error) {
	path, err := parseFileURL(dir)
	if err != nil {
		return nil, err
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
		root: path,
	}, nil
}

// Version returns the configured server info.
func (c *client) Version(_ context.Context) (*driver.Version, error) {
	return c.version, nil
}

// Taken verbatim from http://docs.couchdb.org/en/2.0.0/api/database/common.html
var validDBNameRE = regexp.MustCompile("^[a-z_][a-z0-9_$()+/-]*$")

// AllDBs returns a list of all DBs present in the configured root dir.
func (c *client) AllDBs(ctx context.Context, _ map[string]interface{}) ([]string, error) {
	if c.root == "" {
		return nil, &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "no root path provided"}
	}
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
		filenames = append(filenames, cdb.EscapeID(file.Name()))
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
	return os.Mkdir(filepath.Join(c.root, cdb.EscapeID(dbName)), dirMode)
}

// DBExistsreturns true if the database exists.
func (c *client) DBExists(_ context.Context, dbName string, _ map[string]interface{}) (bool, error) {
	_, err := os.Stat(filepath.Join(c.root, cdb.EscapeID(dbName)))
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
	// FIXME #65: Be safer here about unrecognized files
	return os.RemoveAll(filepath.Join(c.root, cdb.EscapeID(dbName)))
}

func (c *client) DB(_ context.Context, dbName string, _ map[string]interface{}) (driver.DB, error) {
	return c.newDB(dbName)
}

// dbPath returns the full DB path, or an error if the dbpath conflicts with
// the client root path.
func (c *client) dbPath(dbname string) (string, error) {
	if c.root == "" {
		if strings.HasPrefix(dbname, "file://") {
			addr, err := url.Parse(dbname)
			if err != nil {
				return "", &kivik.Error{HTTPStatus: http.StatusBadRequest, Err: err}
			}
			return addr.Path, nil
		}
		return dbname, nil
	}
	if !validDBNameRE.MatchString(dbname) {
		return "", illegalDBName(dbname)
	}
	return filepath.Join(c.root, dbname), nil
}

func (c *client) newDB(dbname string) (*db, error) {
	path, err := c.dbPath(dbname)
	if err != nil {
		return nil, err
	}
	return &db{
		client: c,
		dbName: dbname,
		fs:     c.fs,
		cdb:    cdb.New(path, c.fs),
	}, nil
}
