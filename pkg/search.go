package pkg

import (
	"net/http"
	"fmt"
	"log"
	"io/ioutil"
	"time"
	"strconv"
	"net/url"
	"bytes"
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

func SearchArticles(channelId string, options *SearchArticlesOptions, apiKey string, apiSecret string, baseUrl string) {

	url := fmt.Sprintf("%s/channels/%s/articles", baseUrl, channelId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	query := req.URL.Query()

	options.ApplyToQuery(&query)

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
