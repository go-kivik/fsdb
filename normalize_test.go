package fs

import (
	"encoding/json"
	"os"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/fsdb/filesystem"
	"github.com/go-kivik/fsdb/internal"
)

func TestNormalDocMarshalJSON(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("no attachments", &normalDoc{
		ID:  "foo",
		Rev: internal.Rev{Seq: 1, Sum: "xxx"},
		Data: map[string]interface{}{
			"foo": "bar",
		},
	})
	tests.Add("attachment", func(t *testing.T) interface{} {

		f, err := os.Open("testdata/foo.txt")
		if err != nil {
			t.Fatal(err)
		}

		return &normalDoc{
			ID:  "foo",
			Rev: internal.Rev{Seq: 1, Sum: "xxx"},
			Attachments: attachments{
				"foo.txt": &internal.Attachment{
					ContentType: "text/plain",
					Content:     f,
				},
			},
		}
	})

	tests.Run(t, func(t *testing.T, doc *normalDoc) {
		result, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}

func TestNormalDocUnmarshalJSON(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("no extra fields", `{
        "_id":"foo"
    }`)
	tests.Add("extra fields", `{
        "_id":"foo",
        "foo":"bar"
    }`)
	tests.Add("attachment stub", `{
        "_id":"foo",
        "foo":"bar",
        "_attachments":{
            "foo.txt":{
                "content_type":"text/plain",
                "stub":true
            }
        }
    }`)
	tests.Add("attachment", `{
        "_id":"foo",
        "foo":"bar",
        "_attachments":{
            "foo.txt":{
                "content_type":"text/plain",
                "data":"VGVzdGluZwo="
            }
        }
    }`)

	tests.Run(t, func(t *testing.T, in string) {
		result := new(normalDoc)
		if err := json.Unmarshal([]byte(in), &result); err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}

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
