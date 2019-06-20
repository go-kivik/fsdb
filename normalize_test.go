package fs

import (
	"encoding/json"
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
