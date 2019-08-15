package fs

import (
	"context"
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
