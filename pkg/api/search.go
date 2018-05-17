package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
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

func (c *Client) SearchArticles(options *SearchArticlesOptions) (*SearchArticlesResponse, error) {
	url := fmt.Sprintf("%s/channels/%s/articles", c.BaseURL, c.ChannelID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()

	options.ApplyToQuery(&query)

	req.URL.RawQuery = query.Encode()

	auth, err := c.getAuthorization(http.MethodGet, req.URL.String(), "", ioutil.NopCloser(bytes.NewReader([]byte{})))
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
		return nil, errors.Errorf("SearchArticles - %d", resp.StatusCode)
	}

	var searchArticlesResp SearchArticlesResponse
	err = json.Unmarshal(body, &searchArticlesResp)

	return &searchArticlesResp, err
}
