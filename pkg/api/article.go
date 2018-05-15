package api

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"bytes"
	"mime/multipart"
	"io"
	"strings"
	"net/textproto"
	"log"
	"encoding/json"
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

func CreateArticle(baseUrl string, apiKey string, apiSecret string, channelId string, article io.Reader, metadata *Metadata) {

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
		log.Fatal(err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func UpdateArticle(baseUrl string, apiKey string, apiSecret string, articleId string, article io.Reader, metadata *Metadata) {
	url := fmt.Sprintf("%s/articles/%s", baseUrl, articleId)

	metadataBytes, err := json.Marshal(metadata)

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
		log.Fatal(err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func UpdateArticleMetadata(baseUrl string, apiKey string, apiSecret string, articleId string, metadata *Metadata) {
	url := fmt.Sprintf("%s/articles/%s", baseUrl, articleId)

	metadataBytes, err := json.Marshal(metadata)

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
		log.Fatal(err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func PromoteArticles(baseUrl string, apiKey string, apiSecret string, sectionId string, articleIds []string) {
	url := fmt.Sprintf("%s/sections/%s/promotedArticles", baseUrl, sectionId)

	promotedArticles := PromoteArticlesRequest{}
	if len(articleIds) == 0 {
		articleIds = make([]string, 0)
	}
	promotedArticles.Data.PromotedArticles = articleIds

	bodyBytes, err := json.Marshal(promotedArticles)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))

	req.Header.Set("Authorization", getAuthorization(http.MethodPost, url, apiKey, apiSecret, "", ioutil.NopCloser(bytes.NewReader(bodyBytes))))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func DeleteArticle(baseUrl string, apiKey string, apiSecret string, articleId string) {
	url := fmt.Sprintf("%s/articles/%s", baseUrl, articleId)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", getAuthorization(http.MethodDelete, url, apiKey, apiSecret, "", ioutil.NopCloser(bytes.NewReader([]byte{}))))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
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

	writer.Close()

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", getAuthorization(http.MethodPost, url, apiKey, apiSecret, writer.FormDataContentType(), ioutil.NopCloser(bytes.NewReader(body.Bytes()))))

	return req, err
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}