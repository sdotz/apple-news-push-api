package pkg

import (
	"net/http"
	"fmt"
	"log"
	"io/ioutil"
	"bytes"
)

func ListSections(baseUrl string, apiKey string, apiSecret string, channelId string) {
	url := fmt.Sprintf("%s/channels/%s/sections", baseUrl, channelId)
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
	fmt.Println(string(body))

}
