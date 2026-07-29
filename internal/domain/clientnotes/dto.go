package clientnotes

type CreateNoteRequest struct {
	Note string `json:"note"`
}

type UpdateNoteRequest struct {
	Note string `json:"note"`
}

