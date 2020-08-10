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

package cdb

import (
	"errors"
	"net/http"
	"os"

	"github.com/go-kivik/kivik/v4"
)

var (
	errNotFound         = &kivik.Error{HTTPStatus: http.StatusNotFound, Message: "missing"}
	errUnrecognizedFile = errors.New("unrecognized file")
	errConflict         = &kivik.Error{HTTPStatus: http.StatusConflict, Message: "document update conflict"}
)

// missing transforms a NotExist error into a standard CouchDBesque 'missing'
// error. All other values are passed through unaltered.
func missing(err error) error {
	if !os.IsNotExist(err) {
		return err
	}
	return errNotFound
}

func kerr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, &kivik.Error{}) {
		// Error has already been converted
		return err
	}
	if os.IsNotExist(err) {
		return &kivik.Error{HTTPStatus: http.StatusNotFound, Err: err}
	}
	if os.IsPermission(err) {
		return &kivik.Error{HTTPStatus: http.StatusForbidden, Err: err}
	}
	return err
}
