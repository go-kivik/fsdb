package internal

// RevsInfo is revisions information as presented in the _revs_info key.
type RevsInfo struct {
	Rev    string `json:"rev"`
	Status string `json:"status"`
}

// Revisions is historical revisions data, as presented in the _revisions key.
type Revisions struct {
	Start int64    `json:"start"`
	IDs   []string `json:"ids"`
}
