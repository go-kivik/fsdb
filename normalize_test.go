package fs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/fsdb/internal"
)

func TestAttachmentMarshalJSON(t *testing.T) {
	type tst struct {
		att *attachment
		err string
	}
	tests := testy.NewTable()
	tests.Add("no file", tst{
		att: &attachment{
			ContentType: "text/plain",
		},
		err: "json: error calling MarshalJSON for type *fs.attachment: Content required",
	})
	tests.Add("stub", func(t *testing.T) interface{} {
		f, err := os.Open("testdata/foo.txt")
		if err != nil {
			t.Fatal(err)
		}

		return tst{
			att: &attachment{
				ContentType: "text/plain",
				Content:     f,
				Stub:        true,
			},
		}
	})
	tests.Add("file", func(t *testing.T) interface{} {
		f, err := os.Open("testdata/foo.txt")
		if err != nil {
			t.Fatal(err)
		}

		return tst{
			att: &attachment{
				ContentType: "text/plain",
				Stub:        false,
				Content:     f,
			},
		}
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := json.Marshal(test.att)
		testy.Error(t, test.err, err)
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}

func TestAttachmentUnmarshalJSON_stub(t *testing.T) {
	in := `{
        "content_type": "text/plain",
        "data": "VGVzdGluZwo=",
        "stub": true,
        "digest": "md5-ec4d59b2732f2f153240a8ff746282a6",
        "length": 8
    }`
	result := attachment{}
	if err := json.Unmarshal([]byte(in), &result); err != nil {
		t.Fatal(err)
	}

	expected := attachment{
		ContentType: "text/plain",
		Digest:      "md5-ec4d59b2732f2f153240a8ff746282a6",
		Size:        8,
		Stub:        true,
	}
	if d := testy.DiffInterface(expected, result); d != nil {
		t.Error(d)
	}
}

func TestAttachmentUnmarshalJSON_file(t *testing.T) {
	in := `{
        "content_type": "text/plain",
        "data": "VGVzdGluZwo=",
        "digest": "md5-ec4d59b2732f2f153240a8ff746282a6",
        "length": 8
    }`
	result := attachment{}
	if err := json.Unmarshal([]byte(in), &result); err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadAll(result.Content)
	if err != nil {
		t.Fatal(err)
	}
	defer result.cleanup() // nolint: errcheck
	if d := testy.DiffText("Testing", content); d != nil {
		t.Errorf("content:\n%s", d)
	}
	result.Content = nil
	expected := attachment{
		ContentType: "text/plain",
		Digest:      "md5-ec4d59b2732f2f153240a8ff746282a6",
		Size:        8,
	}
	if d := testy.DiffInterface(expected, result); d != nil {
		t.Error(d)
	}
}

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
				"foo.txt": &attachment{
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
		}
		result, err := db.readDoc(tt.docID, tt.rev)
		testy.StatusError(t, tt.err, tt.status, err)
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}
