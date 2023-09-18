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
	"errors"
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

type metaDoc struct {
	Rev     cdb.RevID `json:"_rev" yaml:"_rev"`
	Deleted bool      `json:"_deleted" yaml:"_deleted"`
}

func (d *db) metadata(docID, ext string) (rev string, deleted bool, err error) {
	f, err := os.Open(d.path(docID))
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return "", false, err
	}
	md := new(metaDoc)
	err = decode.Decode(f, ext, md)
	return md.Rev.String(), md.Deleted, err
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
	return statusError{status: http.StatusBadRequest, error: errors.New("only reserved document ids may start with underscore")}
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
	case kivik.HTTPStatus(err) == http.StatusNotFound:
		// Crate new doc
		doc = d.cdb.NewDocument(docID)
	case err != nil:
		return "", err
	}
	return doc.AddRevision(ctx, rev, opts)
}
