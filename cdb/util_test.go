package cdb

import (
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestEscape(t *testing.T) {
	type tt struct {
		in   string
		want string
	}
	tests := testy.NewTable()
	tests.Add("simple", tt{"simple", "simple"})
	tests.Add("non-ascii", tt{"fóò", "fóò"})
	tests.Add("ddoc", tt{"_design/foo", "_design%2Ffoo"})
	tests.Add("percent", tt{"100%", "100%"})
	tests.Add("escaped slash", tt{"foo%2fbar", "foo%252fbar"})
	tests.Add("empty", tt{"", ""})

	tests.Run(t, func(t *testing.T, tt tt) {
		got := EscapeID(tt.in)
		if got != tt.want {
			t.Errorf("Unexpected escape output: %s", got)
		}
		final := UnescapeID(got)
		if final != tt.in {
			t.Errorf("Unexpected unescape output: %s", final)
		}
	})
}
