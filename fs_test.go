package fs

import (
	"context"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestAllDBs(t *testing.T) {
	type tt struct {
		path     string
		options  map[string]interface{}
		status   int
		err      string
		expected []string
	}
	tests := testy.NewTable()
	tests.Add("testdata", tt{
		path: "testdata",
		expected: []string{
			"compact_extra_atts",
			"compact_noatt",
			"compact_nowinner_noatt",
			"compact_oldrevs",
			"compact_oldrevsatt",
			"compact_split_atts",
			"db_att",
			"db_bar",
			"db_foo",
			"db_put",
			"get_nowinner",
			"get_split_atts",
		},
	})
	tests.Add("No root path", tt{
		path:   "",
		status: http.StatusBadRequest,
		err:    "no root path provided",
	})

	d := &fsDriver{}
	tests.Run(t, func(t *testing.T, tt tt) {
		c, _ := d.NewClient(tt.path)
		result, err := c.AllDBs(context.TODO(), tt.options)
		testy.StatusError(t, tt.err, tt.status, err)
		if d := testy.DiffInterface(tt.expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestClientdbPath(t *testing.T) {
	type tt struct {
		root   string
		dbname string
		status int
		err    string
		path   string
	}
	tests := testy.NewTable()
	tests.Add("normal", tt{
		root:   "/foo/bar",
		dbname: "baz",
		path:   "/foo/bar/baz",
	})
	tests.Add("conflicting absolute paths", tt{
		root:   "foo",
		dbname: "/bar",
		status: http.StatusBadRequest,
		err:    "Name: '/bar'. Only lowercase characters (a-z), digits (0-9), and any of the characters _, $, (, ), +, -, and / are allowed. Must begin with a letter.",
	})
	tests.Add("only db path", tt{
		root:   "",
		dbname: "/foo/bar",
		path:   "/foo/bar",
	})
	tests.Add("invalid file url", tt{
		root:   "",
		dbname: "file:///%xxx",
		status: http.StatusBadRequest,
		err:    `parse file:///%xxx: invalid URL escape "%xx"`,
	})
	tests.Add("file:// url for db", tt{
		root:   "",
		dbname: "file:///foo/bar",
		path:   "/foo/bar",
	})
	tests.Add("file:// url for db with invalid db name", tt{
		root:   "",
		dbname: "file:///foo/bar.baz",
		status: http.StatusBadRequest,
		err:    `Name: 'bar.baz'. Only lowercase characters (a-z), digits (0-9), and any of the characters _, $, (, ), +, -, and / are allowed. Must begin with a letter.`,
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		c := &client{root: tt.root}
		result, err := c.dbPath(tt.dbname)
		testy.StatusError(t, tt.err, tt.status, err)
		if result != tt.path {
			t.Errorf("Unexpected result: %s", result)
		}
	})
}
