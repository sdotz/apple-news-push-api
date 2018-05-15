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

type ReadChannelResponse struct {
	Data struct {
		CreatedAt  time.Time `json:"createdAt"`
		ModifiedAt time.Time `json:"modifiedAt"`
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		ShareURL   string    `json:"shareUrl"`
		Links      Links     `json:"links"`
		Name       string    `json:"name"`
		Website    string    `json:"website"`
	} `json:"data"`
}

func ReadChannel(baseUrl string, apiKey string, apiSecret string, channelId string) (*ReadChannelResponse, error) {
	url := fmt.Sprintf("%s/channels/%s", baseUrl, channelId)
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
		return nil, errors.Errorf("ReadChannel - %d", resp.StatusCode)
	}

	var readChannelResp ReadChannelResponse
	err = json.Unmarshal(body, &readChannelResp)

	return &readChannelResp, err
}
