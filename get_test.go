package fs

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/kivik"
	"github.com/go-kivik/kivik/driver"
)

func TestGet(t *testing.T) {
	type tt struct {
		fs           filesystem.Filesystem
		setup        func(*testing.T, *db)
		final        func(*testing.T, *db)
		path, dbname string
		id           string
		options      map[string]interface{}
		expected     *driver.Document
		status       int
		err          string
	}
	tests := testy.NewTable()
	tests.Add("no id", tt{
		// done
		status: http.StatusBadRequest,
		err:    "no docid specified",
	})
	tests.Add("not found", tt{
		// done
		id:     "foo",
		status: http.StatusNotFound,
		err:    `^missing$`,
	})
	tests.Add("forbidden", func(t *testing.T) interface{} {
		// done
		return tt{
			fs: &filesystem.MockFS{
				OpenFunc: func(_ string) (filesystem.File, error) {
					return nil, &kivik.Error{HTTPStatus: http.StatusForbidden, Err: errors.New("permission denied")}
				},
			},
			id:     "foo",
			status: http.StatusForbidden,
			err:    "permission denied$",
		}
	})
	tests.Add("success, no attachments", tt{
		path:   "testdata",
		dbname: "db.foo",
		id:     "noattach",
		expected: &driver.Document{
			ContentLength: 53,
			Rev:           "1-xxxxxxxxxx",
		},
	})
	tests.Add("success, attachment stub", tt{
		// done
		path:   "testdata",
		dbname: "db.foo",
		id:     "withattach",
		expected: &driver.Document{
			ContentLength: 193,
			Rev:           "2-yyyyyyyyy",
		},
	})
	tests.Add("success, include attachments", tt{
		path:    "testdata",
		dbname:  "db.foo",
		id:      "withattach",
		options: map[string]interface{}{"attachments": true},
		expected: &driver.Document{
			ContentLength: 196,
			Rev:           "2-yyyyyyyyy",
		},
	})
	tests.Add("specify current rev", tt{
		path:    "testdata",
		dbname:  "db.foo",
		id:      "noattach",
		options: map[string]interface{}{"rev": "1-xxxxxxxxxx"},
		expected: &driver.Document{
			ContentLength: 53,
			Rev:           "1-xxxxxxxxxx",
		},
	})
	tests.Add("specify old rev", tt{
		path:    "testdata",
		dbname:  "db.foo",
		id:      "withattach",
		options: map[string]interface{}{"rev": "1-xxxxxxxxxx"},
		expected: &driver.Document{
			ContentLength: 194,
			Rev:           "1-xxxxxxxxxx",
		},
	})
	tests.Add("autorev", tt{
		path:   "testdata",
		dbname: "db.foo",
		id:     "autorev",
		expected: &driver.Document{
			ContentLength: 42,
			Rev:           "6-",
		},
	})
	tests.Add("intrev", tt{
		path:   "testdata",
		dbname: "db.foo",
		id:     "intrev",
		expected: &driver.Document{
			ContentLength: 41,
			Rev:           "6-",
		},
	})
	tests.Add("norev", tt{
		// done
		path:   "testdata",
		dbname: "db.foo",
		id:     "norev",
		expected: &driver.Document{
			ContentLength: 40,
			Rev:           "1-",
		},
	})
	tests.Add("noid", tt{
		path:   "testdata",
		dbname: "db.foo",
		id:     "noid",
		expected: &driver.Document{
			ContentLength: 39,
			Rev:           "6-",
		},
	})
	tests.Add("wrong id", tt{
		path:   "testdata",
		dbname: "db.foo",
		id:     "wrongid",
		expected: &driver.Document{
			ContentLength: 42,
			Rev:           "6-",
		},
	})
	tests.Add("yaml", tt{
		path:   "testdata",
		dbname: "db.foo",
		id:     "yamltest",
		expected: &driver.Document{
			ContentLength: 43,
			Rev:           "3-",
		},
	})
	tests.Add("specify current rev yaml", tt{
		path:    "testdata",
		dbname:  "db.foo",
		id:      "yamltest",
		options: map[string]interface{}{"rev": "3-"},
		expected: &driver.Document{
			ContentLength: 43,
			Rev:           "3-",
		},
	})
	tests.Add("specify old rev yaml", tt{
		path:    "testdata",
		dbname:  "db.foo",
		id:      "yamltest",
		options: map[string]interface{}{"rev": "2-xxx"},
		expected: &driver.Document{
			ContentLength: 46,
			Rev:           "2-xxx",
		},
	})
	tests.Add("specify bogus rev yaml", tt{
		path:    "testdata",
		dbname:  "db.foo",
		id:      "yamltest",
		options: map[string]interface{}{"rev": "1-oink"},
		status:  http.StatusNotFound,
		err:     "missing",
	})
	tests.Add("ddoc yaml", tt{
		path:   "testdata",
		dbname: "db.foo",
		id:     "_design/users",
		expected: &driver.Document{
			ContentLength: 115,
			Rev:           "2-",
		},
	})
	tests.Add("ddoc rev yaml", tt{
		path:    "testdata",
		dbname:  "db.foo",
		id:      "_design/users",
		options: map[string]interface{}{"rev": "2-"},
		expected: &driver.Document{
			ContentLength: 115,
			Rev:           "2-",
		},
	})
	tests.Add("revs", tt{
		path:    "testdata",
		dbname:  "db.foo",
		id:      "wrongid",
		options: map[string]interface{}{"revs": true},
		expected: &driver.Document{
			ContentLength: 93,
			Rev:           "6-",
		},
	})
	tests.Add("revs real", tt{
		path:    "testdata",
		dbname:  "db.foo",
		id:      "noattach",
		options: map[string]interface{}{"revs": true},
		expected: &driver.Document{
			ContentLength: 99,
			Rev:           "1-xxxxxxxxxx",
		},
	})
	tests.Add("note--XkWjFv13acvjJTt-CGJJ8hXlWE", tt{
		path:   "testdata",
		dbname: "db.att",
		id:     "note--XkWjFv13acvjJTt-CGJJ8hXlWE",
		expected: &driver.Document{
			ContentLength: 849,
			Rev:           "1-fbaabe005e0f4e5685a68f857c0777d6",
		},
	})
	tests.Add("note--XkWjFv13acvjJTt-CGJJ8hXlWE + attachments", tt{
		path:    "testdata",
		dbname:  "db.att",
		id:      "note--XkWjFv13acvjJTt-CGJJ8hXlWE",
		options: kivik.Options{"attachments": true},
		expected: &driver.Document{
			ContentLength: 852,
			Rev:           "1-fbaabe005e0f4e5685a68f857c0777d6",
		},
	})
	tests.Add("revs_info=true", tt{
		path:   "testdata",
		dbname: "db.foo",
		id:     "autorev",
		options: kivik.Options{
			"revs_info": true,
		},
		expected: &driver.Document{
			ContentLength: 91,
			Rev:           "6-",
		},
	})
	tests.Add("revs, explicit", tt{
		path:    "testdata",
		dbname:  "db.foo",
		id:      "withrevs",
		options: map[string]interface{}{"revs": true},
		expected: &driver.Document{
			ContentLength: 115,
			Rev:           "8-asdf",
		},
	})
	tests.Add("specify current rev, revs_info=true", tt{
		path:   "testdata",
		dbname: "db.foo",
		id:     "yamltest",
		options: map[string]interface{}{
			"rev":       "3-",
			"revs_info": true,
		},
		expected: &driver.Document{
			ContentLength: 43,
			Rev:           "3-",
		},
	})
	tests.Add("specify conflicting rev, revs_info=true", tt{
		path:   "testdata",
		dbname: "db.foo",
		id:     "yamltest",
		options: map[string]interface{}{
			"rev":       "2-xxx",
			"revs_info": true,
		},
		expected: &driver.Document{
			ContentLength: 46,
			Rev:           "2-xxx",
		},
	})
	tests.Add("specify rev, revs=true", tt{
		path:   "testdata",
		dbname: "db.foo",
		id:     "withrevs",
		options: map[string]interface{}{
			"rev":  "8-asdf",
			"revs": true,
		},
		expected: &driver.Document{
			ContentLength: 47,
			Rev:           "8-asdf",
		},
	})
	tests.Add("interrupted put", tt{
		// This tests a put which was aborted, leaving the attachments in
		// {db}/.{docid}/{rev}/{filename}, while the winning rev is at
		// the friendlier location of {db}/{docid}.{ext}
		path:    "testdata",
		dbname:  "db.foo",
		id:      "abortedput",
		options: map[string]interface{}{"attachments": true},
		expected: &driver.Document{
			ContentLength: 196,
			Rev:           "2-yyyyyyyyy",
		},
	})
	tests.Add("no winner, tied rev", tt{
		// This tests a put which was aborted, leaving the attachments in
		// {db}/.{docid}/{rev}/{filename}, while the winning rev is at
		// the friendlier location of {db}/{docid}.{ext}
		path:   "testdata",
		dbname: "get.nowinner",
		id:     "foo",
		expected: &driver.Document{
			ContentLength: 42,
			Rev:           "1-yyy",
		},
	})
	tests.Add("no winner, greater rev", tt{
		// This tests a put which was aborted, leaving the attachments in
		// {db}/.{docid}/{rev}/{filename}, while the winning rev is at
		// the friendlier location of {db}/{docid}.{ext}
		path:   "testdata",
		dbname: "get.nowinner",
		id:     "bar",
		expected: &driver.Document{
			ContentLength: 39,
			Rev:           "2-yyy",
		},
	})
	tests.Add("atts split between winning and revs dir", tt{
		path:    "testdata",
		dbname:  "get.split_atts",
		id:      "foo",
		options: map[string]interface{}{"attachments": true},
		expected: &driver.Document{
			ContentLength: 292,
			Rev:           "2-zzz",
		},
	})
	tests.Add("atts split between two revs", tt{
		path:    "testdata",
		dbname:  "get.split_atts",
		id:      "bar",
		options: map[string]interface{}{"attachments": true},
		expected: &driver.Document{
			ContentLength: 293,
			Rev:           "2-yyy",
		},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		dir := tt.path
		if dir == "" {
			dir = tempDir(t)
			defer rmdir(t, dir)
		}
		fs := tt.fs
		if fs == nil {
			fs = filesystem.Default()
		}
		db := newDB(&client{root: dir, fs: fs}, tt.dbname)
		if tt.setup != nil {
			tt.setup(t, db)
		}
		doc, err := db.Get(context.Background(), tt.id, tt.options)
		testy.StatusErrorRE(t, tt.err, tt.status, err)
		defer doc.Body.Close() // nolint: errcheck
		if d := testy.DiffAsJSON(testy.Snapshot(t), doc.Body); d != nil {
			t.Errorf("document:\n%s", d)
		}
		if doc.Attachments != nil {
			defer doc.Attachments.Close() // nolint: errcheck
			att := &driver.Attachment{}
			for {
				if err := doc.Attachments.Next(att); err != nil {
					if err == io.EOF {
						break
					}
					t.Fatal(err)
				}
				if d := testy.DiffText(&testy.File{Path: "testdata/" + testy.Stub(t) + "_" + att.Filename}, att.Content); d != nil {
					t.Errorf("Attachment %s content:\n%s", att.Filename, d)
				}
				_ = att.Content.Close()
				att.Content = nil
				if d := testy.DiffAsJSON(&testy.File{Path: "testdata/" + testy.Stub(t) + "_" + att.Filename + "_struct"}, att); d != nil {
					t.Errorf("Attachment %s struct:\n%s", att.Filename, d)
				}
			}
		}
		doc.Body = nil
		doc.Attachments = nil
		if d := testy.DiffInterface(tt.expected, doc); d != nil {
			t.Error(d)
		}
		if tt.final != nil {
			tt.final(t, db)
		}
	})
}
