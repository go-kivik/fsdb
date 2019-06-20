package fs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"
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
		if d := diff.AsJSON(&diff.File{Path: "testdata/" + testy.Stub(t)}, result); d != nil {
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
        "size": 8
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
	if d := diff.Interface(expected, result); d != nil {
		t.Error(d)
	}
}

func TestAttachmentUnmarshalJSON_file(t *testing.T) {
	in := `{
        "content_type": "text/plain",
        "data": "VGVzdGluZwo=",
        "digest": "md5-ec4d59b2732f2f153240a8ff746282a6",
        "size": 8
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
	if d := diff.Text("Testing", content); d != nil {
		t.Errorf("content:\n%s", d)
	}
	result.Content = nil
	expected := attachment{
		ContentType: "text/plain",
		Digest:      "md5-ec4d59b2732f2f153240a8ff746282a6",
		Size:        8,
	}
	if d := diff.Interface(expected, result); d != nil {
		t.Error(d)
	}
}
