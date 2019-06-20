package fs

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/go-kivik/kivik"
)

const attachmentsKey = "_attachments"

func extractAttachments(doc interface{}) (*kivik.Attachments, bool, error) {
	if doc == nil {
		return nil, false, nil
	}
	v := reflect.ValueOf(doc)
	if v.Type().Kind() == reflect.Ptr {
		return extractAttachments(v.Elem().Interface())
	}
	if stdMap, ok := doc.(map[string]interface{}); ok {
		return interfaceToAttachments(stdMap[attachmentsKey])
	}
	if v.Kind() != reflect.Struct {
		return nil, false, nil
	}
	for i := 0; i < v.NumField(); i++ {
		if v.Type().Field(i).Tag.Get("json") == attachmentsKey {
			return interfaceToAttachments(v.Field(i).Interface())
		}
	}
	return nil, false, nil
}

func interfaceToAttachments(i interface{}) (*kivik.Attachments, bool, error) {
	switch t := i.(type) {
	case kivik.Attachments:
		atts := make(kivik.Attachments, len(t))
		for k, v := range t {
			atts[k] = v
			delete(t, k)
		}
		return &atts, true, nil
	case *kivik.Attachments:
		atts := new(kivik.Attachments)
		*atts = *t
		*t = nil
		return atts, true, nil
	case map[string]interface{}:
		return mapToAttachments(t)
	}

	if data, err := json.Marshal(i); err == nil {
		var atts kivik.Attachments
		if err := json.Unmarshal(data, &atts); err != nil {
			return nil, false, &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "bad special document member: _attachments"}
		}
		return &atts, true, nil
	}

	return nil, false, &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "bad special document member: _attachments"}
}

func mapToAttachments(a map[string]interface{}) (*kivik.Attachments, bool, error) {
	atts := make(kivik.Attachments, len(a))
	for filename, d := range a {
		data, ok := d.(map[string]interface{})
		if !ok {
			return nil, false, &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "bad special document member: _attachments"}
		}
		ct, ok := data["content_type"].(string)
		if !ok {
			return nil, false, &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "bad special document member: _attachments"}
		}
		content, ok := data["data"].([]byte)
		if !ok {
			return nil, false, &kivik.Error{HTTPStatus: http.StatusBadRequest, Message: "bad special document member: _attachments"}
		}
		att := &kivik.Attachment{
			ContentType: ct,
			Content:     ioutil.NopCloser(bytes.NewReader(content)),
		}
		atts.Set(filename, att)
	}
	return &atts, true, nil
}
