package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type ContentType string
type CoverArtType string
type MaturityRating string

const (
	ContentTypeJpeg        ContentType = "image/jpeg"
	ContentTypePng         ContentType = "image/png"
	ContentTypeGif         ContentType = "image/gif"
	ContentTypeOctetStream ContentType = "application/octet-stream"
	ContentTypeJson        ContentType = "application/json"
	ContentTypeMultipart   ContentType = "multipart/form-data"

	MaturityRatingKids    = "KIDS"
	MaturityRatingMature  = "MATURE"
	MaturityRatingGeneral = "GENERAL"
)

//Client to communicate to Apple API
type Client struct {
	Client    *http.Client
	APIKey    string
	APISecret string
	BaseURL   string
	ChannelID string
}

type MultipartUploadComponent struct {
	Data     io.Reader
	Name     string
	FileName string
	ContentType
}

type Metadata struct {
	Links struct {
		Sections []string `json:"sections,omitempty"`
	} `json:"links,omitempty"`
	IsSponsored             bool   `json:"isSponsored,omitempty"`
	IsPreview               bool   `json:"isPreview,omitempty"`
	AccessoryText           string `json:"accessoryText,omitempty"`
	MaturityRating          string `json:"maturityRating,omitempty"`
	IsCandidateToBeFeatured bool   `json:"isCandidateToBeFeatured,omitempty"`
	IsHidden                bool   `json:"isHidden,omitempty"`
	Revision                string `json:"revision,omitempty"`
}

type PromoteArticlesRequest struct {
	Data struct {
		PromotedArticles []string `json:"promotedArticles"`
	} `json:"data"`
}

type ReadArticleResponse struct {
	Data struct {
		CreatedAt               time.Time     `json:"createdAt"`
		ModifiedAt              time.Time     `json:"modifiedAt"`
		ID                      string        `json:"id"`
		Type                    string        `json:"type"`
		ShareURL                string        `json:"shareUrl"`
		Links                   Links         `json:"links"`
		Document                interface{}   `json:"document"`
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

func (c *Client) ReadArticle(articleId string) (*ReadArticleResponse, error) {
	url := fmt.Sprintf("%s/articles/%s", c.BaseURL, articleId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	auth, err := c.getAuthorization(http.MethodGet, url, "", ioutil.NopCloser(bytes.NewReader([]byte{})))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("ReadArticle - %d", resp.StatusCode)
	}

	var readArticleResp ReadArticleResponse
	err = json.Unmarshal(body, &readArticleResp)

	if err != nil {
		return nil, err
	}

	return &readArticleResp, nil
}

func (c *Client) CreateArticle(article io.Reader, metadata *Metadata) error {
	url := fmt.Sprintf("%s/channels/%s/articles", c.BaseURL, c.ChannelID)

	req, err := c.prepareMultipartRequest(
		[]MultipartUploadComponent{
			{
				Data:        article,
				Name:        "article.json",
				FileName:    "article.json",
				ContentType: ContentTypeJson,
			},
		},
		url,
	)

	if err != nil {
		return err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return resp.Body.Close()
}

func (c *Client) UpdateArticle(articleId string, article io.Reader, metadata *Metadata) error {
	url := fmt.Sprintf("%s/articles/%s", c.BaseURL, articleId)

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	parts := []MultipartUploadComponent{
		{
			Data:        bytes.NewReader(metadataBytes),
			Name:        "metadata",
			ContentType: ContentTypeJson,
		},
		{
			Data:        article,
			Name:        "article.json",
			FileName:    "article.json",
			ContentType: ContentTypeJson,
		},
	}

	req, err := c.prepareMultipartRequest(
		parts,
		url,
	)

	if err != nil {
		return err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return resp.Body.Close()
}

func (c *Client) UpdateArticleMetadata(articleId string, metadata *Metadata) error {
	url := fmt.Sprintf("%s/articles/%s", c.BaseURL, articleId)

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	req, err := c.prepareMultipartRequest(
		[]MultipartUploadComponent{
			{
				Data:        bytes.NewReader(metadataBytes),
				Name:        "metadata",
				ContentType: ContentTypeJson,
			},
		},
		url,
	)

	if err != nil {
		return err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return resp.Body.Close()
}

func (c *Client) PromoteArticles(sectionId string, articleIds []string) error {
	url := fmt.Sprintf("%s/sections/%s/promotedArticles", c.BaseURL, sectionId)

	promotedArticles := PromoteArticlesRequest{}
	if len(articleIds) == 0 {
		articleIds = make([]string, 0)
	}
	promotedArticles.Data.PromotedArticles = articleIds

	bodyBytes, err := json.Marshal(promotedArticles)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))

	auth, err := c.getAuthorization(http.MethodPost, url, "", ioutil.NopCloser(bytes.NewReader([]byte{})))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return resp.Body.Close()
}

func (c *Client) DeleteArticle(articleId string) error {
	url := fmt.Sprintf("%s/articles/%s", c.BaseURL, articleId)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	auth, err := c.getAuthorization(http.MethodDelete, url, "", ioutil.NopCloser(bytes.NewReader([]byte{})))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return resp.Body.Close()
}

func (c *Client) prepareMultipartRequest(parts []MultipartUploadComponent, url string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, v := range parts {
		h := make(textproto.MIMEHeader)
		contentDispositionHeader := fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(v.Name))
		if len(v.FileName) > 0 {
			contentDispositionHeader = contentDispositionHeader + fmt.Sprintf(`; filename="%s"`, v.FileName)
		}
		h.Set("Content-Disposition", contentDispositionHeader)
		h.Set("Content-Type", string(v.ContentType))
		part, err := writer.CreatePart(h)
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(part, v.Data)
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	auth, err := c.getAuthorization(http.MethodPost, url, writer.FormDataContentType(), ioutil.NopCloser(bytes.NewReader(body.Bytes())))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)

	return req, err
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}
