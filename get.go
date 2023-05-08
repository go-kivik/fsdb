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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

// TODO:
// - atts_since
// - conflicts
// - deleted_conflicts
// - latest
// - local_seq
// - meta
// - open_revs
func (d *db) Get(ctx context.Context, docID string, opts map[string]interface{}) (*driver.Document, error) {
	if docID == "" {
		return nil, &kivik.Error{Status: http.StatusBadRequest, Message: "no docid specified"}
	}
	doc, err := d.cdb.OpenDocID(docID, opts)
	if err != nil {
		return nil, err
	}
	doc.Options = opts
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(doc); err != nil {
		return nil, err
	}
	attsIter, err := doc.Revisions[0].AttachmentsIterator()
	if err != nil {
		return nil, err
	}
	return &driver.Document{
		Rev:         doc.Revisions[0].Rev.String(),
		Body:        io.NopCloser(buf),
		Attachments: attsIter,
	}, nil
}
