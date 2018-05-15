package api

import (
	"log"
	"os"
)

var (
	stderr = log.New(os.Stderr, "", 0)
	stdout = log.New(os.Stdout, "", 0)
)

type Links struct {
	Channel        string   `json:"channel,omitempty"`
	Next           string   `json:"next,omitempty"`
	Self           string   `json:"self"`
	DefaultSection string   `json:"defaultSection,omitempty"`
	Sections       []string `json:"sections,omitempty"`
}
