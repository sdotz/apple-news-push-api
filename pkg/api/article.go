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

func ReadArticle(baseUrl string, apiKey string, apiSecret string, articleId string) (*ReadArticleResponse, error) {
	url := fmt.Sprintf("%s/articles/%s", baseUrl, articleId)
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
		stderr.Printf("ReadArticle - %d\n", resp.StatusCode)
		stderr.Println(string(body))
		return nil, errors.Errorf("ReadArticle - %d", resp.StatusCode)
	}

	var readArticleResp ReadArticleResponse
	err = json.Unmarshal(body, &readArticleResp)

	if err != nil {
		return nil, err
	}

	return &readArticleResp, nil
}

func CreateArticle(baseUrl string, apiKey string, apiSecret string, channelId string, article io.Reader, metadata *Metadata) error {

	url := fmt.Sprintf("%s/channels/%s/articles", baseUrl, channelId)

	req, err := prepareMultipartRequest(
		[]MultipartUploadComponent{
			{
				Data:        article,
				Name:        "article.json",
				FileName:    "article.json",
				ContentType: ContentTypeJson,
			},
		},
		url,
		apiKey,
		apiSecret,
	)

	if err != nil {
		return err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return resp.Body.Close()
}

func UpdateArticle(baseUrl string, apiKey string, apiSecret string, articleId string, article io.Reader, metadata *Metadata) error {
	url := fmt.Sprintf("%s/articles/%s", baseUrl, articleId)

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

	req, err := prepareMultipartRequest(
		parts,
		url,
		apiKey,
		apiSecret,
	)

	if err != nil {
		return err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return resp.Body.Close()
}

func UpdateArticleMetadata(baseUrl string, apiKey string, apiSecret string, articleId string, metadata *Metadata) error {
	url := fmt.Sprintf("%s/articles/%s", baseUrl, articleId)

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	req, err := prepareMultipartRequest(
		[]MultipartUploadComponent{
			{
				Data:        bytes.NewReader(metadataBytes),
				Name:        "metadata",
				ContentType: ContentTypeJson,
			},
		},
		url,
		apiKey,
		apiSecret,
	)

	if err != nil {
		return err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return resp.Body.Close()
}

func PromoteArticles(baseUrl string, apiKey string, apiSecret string, sectionId string, articleIds []string) error {
	url := fmt.Sprintf("%s/sections/%s/promotedArticles", baseUrl, sectionId)

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

	auth, err := getAuthorization(http.MethodPost, url, apiKey, apiSecret, "", ioutil.NopCloser(bytes.NewReader([]byte{})))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return resp.Body.Close()
}

func DeleteArticle(baseUrl string, apiKey string, apiSecret string, articleId string) error {
	url := fmt.Sprintf("%s/articles/%s", baseUrl, articleId)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	auth, err := getAuthorization(http.MethodDelete, url, apiKey, apiSecret, "", ioutil.NopCloser(bytes.NewReader([]byte{})))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return resp.Body.Close()
}

func prepareMultipartRequest(parts []MultipartUploadComponent, url string, apiKey string, apiSecret string) (*http.Request, error) {
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

	auth, err := getAuthorization(http.MethodPost, url, apiKey, apiSecret, writer.FormDataContentType(), ioutil.NopCloser(bytes.NewReader(body.Bytes())))
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
