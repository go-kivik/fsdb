package fs

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"

	"github.com/go-kivik/kivik"
)

type customAttachments struct{}

func (a customAttachments) MarshalJSON() ([]byte, error) {
	return []byte(`{"foo.txt":{"content_type":"text/html","data":"PGgxPnRlc3Rpbmc8L2gxPg=="}}`), nil
}

func TestExtractAttachments(t *testing.T) {
	type tst struct {
		doc      interface{}
		expected *kivik.Attachments
		status   int
		err      string
	}
	tests := testy.NewTable()
	tests.Add("nil doc", tst{
		doc:      nil,
		expected: nil,
	})
	tests.Add("doc is pointer", tst{
		doc:      &struct{}{},
		expected: nil,
	})
	tests.Add("No attachments", tst{
		doc:      map[string]string{"foo": "bar"},
		expected: nil,
	})
	tests.Add("*kivik.Attachments in map", tst{
		doc: map[string]interface{}{
			"_attachments": &kivik.Attachments{
				"foo.txt": &kivik.Attachment{
					ContentType: "text/plain",
					Content:     ioutil.NopCloser(strings.NewReader("testing")),
				},
			},
		},
		expected: &kivik.Attachments{
			"foo.txt": &kivik.Attachment{
				ContentType: "text/plain",
				Content:     ioutil.NopCloser(strings.NewReader("testing")),
			},
		},
	})
	tests.Add("kivik.Attachments in map", tst{
		doc: map[string]interface{}{
			"_attachments": kivik.Attachments{
				"foo.txt": &kivik.Attachment{
					ContentType: "text/plain",
					Content:     ioutil.NopCloser(strings.NewReader("testing")),
				},
			},
		},
		expected: &kivik.Attachments{
			"foo.txt": &kivik.Attachment{
				ContentType: "text/plain",
				Content:     ioutil.NopCloser(strings.NewReader("testing")),
			},
		},
	})
	tests.Add("*kivik.Attachments in struct", tst{
		doc: struct {
			Att *kivik.Attachments `json:"_attachments"`
		}{
			Att: &kivik.Attachments{
				"foo.txt": &kivik.Attachment{
					ContentType: "text/plain",
					Content:     ioutil.NopCloser(strings.NewReader("testing")),
				},
			},
		},
		expected: &kivik.Attachments{
			"foo.txt": &kivik.Attachment{
				ContentType: "text/plain",
				Content:     ioutil.NopCloser(strings.NewReader("testing")),
			},
		},
	})
	tests.Add("attachments as standard map", tst{
		doc: map[string]interface{}{
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"content_type": "text/plain",
					"data":         []byte("testing"),
				},
			},
		},
		expected: &kivik.Attachments{
			"foo.txt": &kivik.Attachment{
				ContentType: "text/plain",
				Content:     ioutil.NopCloser(strings.NewReader("testing")),
			},
		},
	})
	tests.Add("attachments with custom marshaler", tst{
		doc: map[string]interface{}{
			"_attachments": customAttachments{},
		},
		expected: &kivik.Attachments{
			"foo.txt": &kivik.Attachment{
				ContentType: "text/html",
				Content:     ioutil.NopCloser(strings.NewReader("<h1>testing</h1>")),
			},
		},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := extractAttachments(test.doc)
		testy.StatusError(t, test.err, test.status, err)
		if d := diff.AsJSON(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}
