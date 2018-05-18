package api

import (
	"net/http"
)

type Links struct {
	Channel        string   `json:"channel,omitempty"`
	Next           string   `json:"next,omitempty"`
	Self           string   `json:"self"`
	DefaultSection string   `json:"defaultSection,omitempty"`
	Sections       []string `json:"sections,omitempty"`
}

func NewClient(httpClient *http.Client, key, secret, url, channelID string) *Client {
	return &Client{
		Client:    httpClient,
		APIKey:    key,
		APISecret: secret,
		BaseURL:   url,
		ChannelID: channelID,
	}
}
