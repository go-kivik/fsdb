// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package fs

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-kivik/fsdb/v4/cdb"
	"github.com/go-kivik/fsdb/v4/cdb/decode"
	"github.com/go-kivik/kivik/v4"
)

func filename2id(filename string) (string, error) {
	return url.PathUnescape(filename)
}

type revDoc struct {
	Rev cdb.RevID `json:"_rev" yaml:"_rev"`
}

func (d *db) currentRev(docID, ext string) (string, error) {
	f, err := os.Open(d.path(docID))
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return "", err
	}
	rd := new(revDoc)
	err = decode.Decode(f, ext, rd)
	return rd.Rev.String(), err
}

var reservedPrefixes = []string{"_local/", "_design/"}

func validateID(id string) error {
	if id[0] != '_' {
		return nil
	}
	for _, prefix := range reservedPrefixes {
		if strings.HasPrefix(id, prefix) && len(id) > len(prefix) {
			return nil
		}
	}
	return &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "only reserved document ids may start with underscore"}
}

/*
TODO:
URL query params:
batch
new_edits

output_format?

X-Couch-Full-Commit header/option
*/

func (d *db) Put(ctx context.Context, docID string, i interface{}, opts map[string]interface{}) (string, error) {
	if err := validateID(docID); err != nil {
		return "", err
	}
	rev, err := d.cdb.NewRevision(i)
	if err != nil {
		return "", err
	}
	doc, err := d.cdb.OpenDocID(docID, opts)
	switch {
	case kivik.StatusCode(err) == http.StatusNotFound:
		// Crate new doc
		doc = d.cdb.NewDocument(docID)
	case err != nil:
		return "", err
	}
	return doc.AddRevision(ctx, rev, opts)
}
