package pkg

import (
	"net/http"
	"fmt"
	"log"
	"io/ioutil"
)

func ReadArticle(articleId string, apiKey string, apiSecret string, baseUrl string) {
	url := fmt.Sprintf("%s/articles/%s", baseUrl, articleId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", getAuthorization(http.MethodGet, url, apiKey, apiSecret, "", ""))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

}

func ReadChannel(channelId string, apiKey string, apiSecret string, baseUrl string) {
	url := fmt.Sprintf("%s/channels/%s", baseUrl, channelId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", getAuthorization(http.MethodGet, url, apiKey, apiSecret, "", ""))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

}

func ReadSection(sectionId string, apiKey string, apiSecret string, baseUrl string) {
	url := fmt.Sprintf("%s/sections/%s", baseUrl, sectionId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", getAuthorization(http.MethodGet, url, apiKey, apiSecret, "", ""))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}
