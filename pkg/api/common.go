package api

type Links struct {
	Channel        string   `json:"channel,omitempty"`
	Next           string   `json:"next,omitempty"`
	Self           string   `json:"self"`
	DefaultSection string   `json:"defaultSection,omitempty"`
	Sections       []string `json:"sections,omitempty"`
}
