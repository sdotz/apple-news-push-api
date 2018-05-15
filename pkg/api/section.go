package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type ReadSectionResponse struct {
	Data struct {
		CreatedAt  time.Time `json:"createdAt"`
		ModifiedAt time.Time `json:"modifiedAt"`
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		ShareURL   string    `json:"shareUrl"`
		Links      Links     `json:"links"`
		Name       string    `json:"name"`
		IsDefault  bool      `json:"isDefault"`
	} `json:"data"`
}

type ListSectionsResponse struct {
	Data []struct {
		CreatedAt  time.Time `json:"createdAt"`
		ModifiedAt time.Time `json:"modifiedAt"`
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		ShareURL   string    `json:"shareUrl"`
		Links      Links     `json:"links"`
		Name       string    `json:"name"`
		IsDefault  bool      `json:"isDefault"`
	} `json:"data"`
}

func ReadSection(baseUrl string, apiKey string, apiSecret string, sectionId string) (*ReadSectionResponse, error) {
	url := fmt.Sprintf("%s/sections/%s", baseUrl, sectionId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	auth, err := getAuthorization(http.MethodGet, url, apiKey, apiSecret, "", ioutil.NopCloser(bytes.NewReader([]byte{})))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("ReadSection - %d", resp.StatusCode)
	}

	var readSectionResp ReadSectionResponse
	err = json.Unmarshal(body, &readSectionResp)

	if err != nil {
		return nil, err
	}

	return &readSectionResp, nil
}

func ListSections(baseUrl string, apiKey string, apiSecret string, channelId string) (*ListSectionsResponse, error) {
	url := fmt.Sprintf("%s/channels/%s/sections", baseUrl, channelId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	auth, err := getAuthorization(http.MethodGet, url, apiKey, apiSecret, "", ioutil.NopCloser(bytes.NewReader([]byte{})))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("ListSections - %d", resp.StatusCode)
	}

	var listSectionsResp ListSectionsResponse
	err = json.Unmarshal(body, &listSectionsResp)

	return &listSectionsResp, err

}
