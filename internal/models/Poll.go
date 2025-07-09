package models

type PollOption struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type Poll struct {
	ID          int           `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	UserID      int           `json:"user_id"`
	Options     []*PollOption `json:"options"`
}

type Vote struct {
	ID       int `json:"id"`
	OptionID int `json:"option_id"`
	UserID   int `json:"user_id"`
}
