package fs

import (
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/fsdb/filesystem"
)

func TestNormalizeDoc(t *testing.T) {
	type tst struct {
		doc    interface{}
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("simple doc", tst{
		doc: map[string]string{"_id": "foo", "foo": "bar"},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := normalizeDoc(test.doc)
		testy.StatusError(t, test.err, test.status, err)
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}

func TestReadDoc(t *testing.T) {
	type tt struct {
		root, dbname, docID, rev string
		status                   int
		err                      string
	}
	tests := testy.NewTable()
	tests.Add("json", tt{
		root:   "testdata",
		dbname: "db.foo",
		docID:  "noattach",
	})
	tests.Add("yaml", tt{
		root:   "testdata",
		dbname: "db.foo",
		docID:  "yamltest",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		db := &db{
			client: &client{root: tt.root},
			dbName: tt.dbname,
			fs:     filesystem.Default(),
		}
		result, err := db.readDoc(tt.docID, tt.rev)
		testy.StatusError(t, tt.err, tt.status, err)
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}
