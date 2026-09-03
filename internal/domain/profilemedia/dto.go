package profilemedia

type CreateMediaRequest struct {
	MediaType MediaType `json:"media_type"`
	URL       string    `json:"url"`
	Caption   string    `json:"caption"`
	SortOrder int       `json:"sort_order"`
}

type UpdateMediaRequest struct {
	Caption   string `json:"caption"`
	SortOrder int    `json:"sort_order"`
}
