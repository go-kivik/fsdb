package internal

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestAttachmentMarshalJSON(t *testing.T) {
	type tst struct {
		att *Attachment
		err string
	}
	tests := testy.NewTable()
	tests.Add("no file", tst{
		att: &Attachment{
			ContentType: "text/plain",
		},
		err: "json: error calling MarshalJSON for type *internal.Attachment: Content required",
	})
	tests.Add("stub", func(t *testing.T) interface{} {
		f, err := os.Open("testdata/foo.txt")
		if err != nil {
			t.Fatal(err)
		}

		return tst{
			att: &Attachment{
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
			att: &Attachment{
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
	result := Attachment{}
	if err := json.Unmarshal([]byte(in), &result); err != nil {
		t.Fatal(err)
	}

	expected := Attachment{
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
	result := Attachment{}
	if err := json.Unmarshal([]byte(in), &result); err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadAll(result.Content)
	if err != nil {
		t.Fatal(err)
	}
	defer result.Cleanup() // nolint: errcheck
	if d := testy.DiffText("Testing", content); d != nil {
		t.Errorf("content:\n%s", d)
	}
	result.Content = nil
	expected := Attachment{
		ContentType: "text/plain",
		Digest:      "md5-ec4d59b2732f2f153240a8ff746282a6",
		Size:        8,
	}
	if d := testy.DiffInterface(expected, result); d != nil {
		t.Error(d)
	}
}
