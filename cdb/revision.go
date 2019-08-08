package cdb

import (
	"encoding/json"
)

// RevMeta is the metadata stored in reach revision.
type RevMeta struct {
	Rev         RevID                  `json:"_rev" yaml:"_rev"`
	Deleted     *bool                  `json:"_deleted,omitempty" yaml:"_deleted,omitempty"`
	Attachments map[string]*Attachment `json:"_attachments,omitempty" yaml:"_attachments,omitempty"`
	RevHistory  *RevHistory            `json:"_revisions,omitempty" yaml:"_revisions,omitempty"`
}

// Revision is a specific instance of a document.
type Revision struct {
	RevMeta

	// Data is the normal payload
	Data map[string]interface{} `json:"-" yaml:"-"`
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (r *Revision) UnmarshalJSON(p []byte) error {
	if err := json.Unmarshal(p, &r.RevMeta); err != nil {
		return err
	}
	if err := json.Unmarshal(p, &r.Data); err != nil {
		return err
	}
	for key := range reservedKeys {
		delete(r.Data, key)
	}
	return nil
}

// UnmarshalYAML satisfies the yaml.Unmarshaler interface.
func (r *Revision) UnmarshalYAML(u func(interface{}) error) error {
	if err := u(&r.RevMeta); err != nil {
		return err
	}
	if err := u(&r.Data); err != nil {
		return err
	}
	for key := range reservedKeys {
		delete(r.Data, key)
	}
	return nil
}

// MarshalJSON satisfies the json.Marshaler interface
func (r *Revision) MarshalJSON() ([]byte, error) {
	metaJSON, err := json.Marshal(r.RevMeta)
	if err != nil {
		return nil, err
	}
	if len(r.Data) == 0 {
		return metaJSON, nil
	}
	dataJSON, err := json.Marshal(r.Data)
	if err != nil {
		return nil, err
	}
	result := make([]byte, 0, len(dataJSON)+len(metaJSON)-1)
	result = append(result, metaJSON[:len(metaJSON)-1]...)
	result = append(result, ',')
	result = append(result, dataJSON[1:]...)
	return result, nil
}
