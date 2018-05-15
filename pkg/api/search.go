package api

import (
	"net/http"
	"fmt"
	"log"
	"io/ioutil"
	"time"
	"strconv"
	"net/url"
	"bytes"
	"github.com/pkg/errors"
	"encoding/json"
)

type ResultPage struct {
	Results []SearchResult
}

type SearchResult struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}

const SORTDIR_DESC = "DESC"
const SORTDIR_ASC = "ASC"

type SearchArticlesOptions struct {
	SortDir   string
	FromDate  *time.Time
	ToDate    *time.Time
	PageSize  int
	PageToken string
}

type SearchArticlesResponse struct {
	Data []struct {
		CreatedAt               time.Time     `json:"createdAt"`
		ModifiedAt              time.Time     `json:"modifiedAt"`
		ID                      string        `json:"id"`
		Type                    string        `json:"type"`
		ShareURL                string        `json:"shareUrl"`
		Links                   Links         `json:"links"`
		Revision                string        `json:"revision"`
		State                   string        `json:"state"`
		AccessoryText           string        `json:"accessoryText"`
		Title                   string        `json:"title"`
		MaturityRating          string        `json:"maturityRating"`
		Warnings                []interface{} `json:"warnings"`
		IsCandidateToBeFeatured bool          `json:"isCandidateToBeFeatured"`
		IsSponsored             bool          `json:"isSponsored"`
		IsPreview               bool          `json:"isPreview"`
		IsDevelopingStory       bool          `json:"isDevelopingStory"`
		IsHidden                bool          `json:"isHidden"`
	} `json:"data"`
	Links Links `json:"links"`
}

func DefaultSearchArticlesOptions() *SearchArticlesOptions {
	return &SearchArticlesOptions{
		SortDir:   SORTDIR_DESC,
		FromDate:  nil,
		ToDate:    nil,
		PageSize:  10,
		PageToken: "",
	}
}

func (options *SearchArticlesOptions) ApplyToQuery(query *url.Values) {
	query.Add("sortDir", options.SortDir)
	if options.FromDate != nil {
		query.Add("fromDate", options.FromDate.Format(time.RFC3339))
	}

	if options.ToDate != nil {
		query.Add("toDate", options.ToDate.Format(time.RFC3339))
	}

	query.Add("pageSize", strconv.Itoa(options.PageSize))
	query.Add("pageToken", options.PageToken)
}

func SearchArticles(baseUrl string, apiKey string, apiSecret string, channelId string, options *SearchArticlesOptions) (*SearchArticlesResponse, error) {
	url := fmt.Sprintf("%s/channels/%s/articles", baseUrl, channelId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	query := req.URL.Query()

	options.ApplyToQuery(&query)

	req.URL.RawQuery = query.Encode()

	req.Header.Set("Authorization", getAuthorization(http.MethodGet, req.URL.String(), apiKey, apiSecret, "", ioutil.NopCloser(bytes.NewReader([]byte{}))))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		stderr.Printf("SearchArticles - %d\n", resp.StatusCode)
		stderr.Println(string(body))
		return nil, errors.Errorf("SearchArticles - %d", resp.StatusCode)
	}

	var searchArticlesResp SearchArticlesResponse
	err = json.Unmarshal(body, &searchArticlesResp)

	if err != nil {
		log.Fatal(err)
	}

	return &searchArticlesResp, nil
}
