package fs

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-kivik/kivik"
)

type revDoc struct {
	Rev string `json:"_rev"`
}

func calculateRev(doc map[string]interface{}) string {
	seq := 1
	if rev, ok := doc["_rev"].(string); ok {
		seq, _ = strconv.Atoi(strings.SplitN(rev, "-", 2)[0])
	}
	token := fmt.Sprintf("%d %s", seq, doc["_id"].(string))
	return fmt.Sprintf("%d-%x", seq, md5.Sum([]byte(token)))
}

func (d *db) currentRev(docID string) (string, error) {
	f, err := os.Open(d.path(docID))
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return "", err
	}
	var rd revDoc
	err = json.NewDecoder(f).Decode(&rd)
	return rd.Rev, err
}

func compareRevs(doc, opts map[string]interface{}, currev string) error {
	optsrev, _ := opts["rev"].(string)
	docrev, _ := doc["_rev"].(string)
	if optsrev != "" && docrev != "" && optsrev != docrev {
		return &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "document rev from request body and query string have different values"}
	}
	if currev == "" {
		return nil
	}
	if optsrev != "" && optsrev != currev {
		return &kivik.Error{HTTPStatus: http.StatusConflict, Message: "document update conflict"}
	}
	if docrev != "" && docrev != currev {
		return &kivik.Error{HTTPStatus: http.StatusConflict, Message: "document update conflict"}
	}
	return nil
}

/*
File naming strategy:
Current rev lives under {db}/{docid}
Historical revs live under {db}/.{docid}/{rev}
*/
func (d *db) Put(_ context.Context, docID string, doc interface{}, opts map[string]interface{}) (string, error) {
	currev, err := d.currentRev(docID)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	var obj map[string]interface{}
	_ = json.Unmarshal(data, &obj)
	if err := compareRevs(obj, opts, currev); err != nil {
		return "", err
	}
	obj["_id"] = docID
	rev := calculateRev(obj)
	obj["_rev"] = rev

	f, err := os.OpenFile(d.path(docID), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(obj); err != nil {
		return "", err
	}
	return rev, nil
}
