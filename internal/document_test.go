package internal

import (
	"encoding/json"
	"os"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestDocumentMarshalJSON(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("no attachments", &Document{
		DocMeta: DocMeta{
			ID:  "foo",
			Rev: Rev{Seq: 1, Sum: "xxx"},
		},
		Data: map[string]interface{}{
			"foo": "bar",
		},
	})
	tests.Add("attachment", func(t *testing.T) interface{} {

		f, err := os.Open("testdata/foo.txt")
		if err != nil {
			t.Fatal(err)
		}

		return &Document{
			DocMeta: DocMeta{
				ID:  "foo",
				Rev: Rev{Seq: 1, Sum: "xxx"},
				Attachments: Attachments{
					"foo.txt": &Attachment{
						ContentType: "text/plain",
						Content:     f,
					},
				},
			},
		}
	})

	tests.Run(t, func(t *testing.T, doc *Document) {
		result, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}

func TestDocumentUnmarshalJSON(t *testing.T) {
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
		result := new(Document)
		if err := json.Unmarshal([]byte(in), &result); err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}
