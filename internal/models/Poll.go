package models

type PollOption struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type Poll struct {
	ID          int           `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Options     []*PollOption `json:"options"`
}
