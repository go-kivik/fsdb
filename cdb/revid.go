package cdb

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// RevID is a CouchDB document revision identifier.
type RevID struct {
	Seq      int64
	Sum      string
	original string
}

func (r *RevID) Changed() bool {
	return r.String() != r.original
}

func (r *RevID) UnmarshalText(p []byte) error {
	r.original = string(p)
	if bytes.Contains(p, []byte("-")) {
		parts := bytes.SplitN(p, []byte("-"), 2)
		seq, err := strconv.ParseInt(string(parts[0]), 10, 64)
		if err != nil {
			return err
		}
		r.Seq = seq
		if len(parts) > 1 {
			r.Sum = string(parts[1])
		}
		return nil
	}
	r.Sum = ""
	seq, err := strconv.ParseInt(string(p), 10, 64)
	if err != nil {
		return err
	}
	r.Seq = seq
	return nil
}

func (r *RevID) UnmarshalJSON(p []byte) error {
	if p[0] == '"' {
		var str string
		if e := json.Unmarshal(p, &str); e != nil {
			return e
		}
		r.original = str
		parts := strings.SplitN(str, "-", 2)
		seq, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return err
		}
		r.Seq = seq
		if len(parts) > 1 {
			r.Sum = parts[1]
		}
		return nil
	}
	r.original = string(p)
	r.Sum = ""
	return json.Unmarshal(p, &r.Seq)
}

func (r RevID) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r RevID) String() string {
	if r.Seq == 0 {
		return ""
	}
	return fmt.Sprintf("%d-%s", r.Seq, r.Sum)
}

func (r *RevID) IsZero() bool {
	return r.Seq == 0
}

func (r *RevID) Increment(payload ...string) {
	r.Seq++
	if len(payload) == 0 {
		r.Sum = ""
		return
	}
	data := strings.Join(payload, "")
	r.Sum = fmt.Sprintf("%x", md5.Sum([]byte(data)))
}
