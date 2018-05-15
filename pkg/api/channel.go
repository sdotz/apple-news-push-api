package api

import (
	"time"
	"fmt"
	"net/http"
	"io/ioutil"
	"bytes"
	"log"
	"github.com/pkg/errors"
	"encoding/json"
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
		log.Fatal(err)
	}
	req.Header.Set("Authorization", getAuthorization(http.MethodGet, url, apiKey, apiSecret, "", ioutil.NopCloser(bytes.NewReader([]byte{}))))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		stderr.Printf("ReadChannel - %d\n", resp.StatusCode)
		stderr.Println(string(body))
		return nil, errors.Errorf("ReadChannel - %d", resp.StatusCode)
	}

	var readChannelResp ReadChannelResponse
	err = json.Unmarshal(body, &readChannelResp)

	if err != nil {
		log.Fatal(err)
	}

	return &readChannelResp, nil
}
