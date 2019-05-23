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

	"os"
	"path/filepath"
	"regexp"

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
	ContentTypeHtml        ContentType = "text/html"
	ContentTypeCss         ContentType = "text/css"
	ContentTypeJs          ContentType = "text/javascript"
	ContentTypeManifest    ContentType = "text/cache-manifest"
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

type Data struct {
	Links                   Links  `json:"links,omitempty"`
	IsSponsored             bool   `json:"isSponsored,omitempty"`
	IsPreview               bool   `json:"isPreview,omitempty"`
	AccessoryText           string `json:"accessoryText,omitempty"`
	MaturityRating          string `json:"maturityRating,omitempty"`
	IsCandidateToBeFeatured bool   `json:"isCandidateToBeFeatured,omitempty"`
	IsDevelopingStory       bool   `json:"isDevelopingStory,omitempty"`
	IsHidden                bool   `json:"isHidden,omitempty"`
	Revision                string `json:"revision,omitempty"`
}

type Metadata struct {
	Data Data `json:"data"`
}

type Meta struct {
	Throttling struct {
		IsThrottled             bool `json:"isThrottled,omitempty"`
		QuotaAvailable          int  `json:"quotaAvailable,omitempty"`
		QueueSize               int  `json:"queueSize,omitempty"`
		EstimatedDelayInSeconds int  `json:"estimatedDelayInSeconds,omitempty"`
	} `json:"throttling,omitempty"`
}

type PromoteArticlesRequest struct {
	Data struct {
		PromotedArticles []string `json:"promotedArticles"`
	} `json:"data"`
}

type PromoteArticlesResponse struct {
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
	Meta Meta `json:"meta,omitempty"`
}

type BundleComponent struct {
	Data io.Reader
	Name string
	Ext  string
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

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("%s returned a %d", url, resp.StatusCode)
	}

	var readArticleResp ReadArticleResponse
	err = json.Unmarshal(body, &readArticleResp)

	if err != nil {
		return nil, err
	}

	return &readArticleResp, nil
}

func GetBundleComponents(articleJson io.Reader, bundleBasePath string) ([]MultipartUploadComponent, error) {
	var matches [][]string
	var bundleComponents []MultipartUploadComponent

	re := regexp.MustCompile(`"bundle:\/\/(.*?)"`)

	articleBytes, err := ioutil.ReadAll(articleJson)
	if err != nil {
		return nil, err
	}

	matches = re.FindAllStringSubmatch(string(articleBytes), -1)

	if len(matches) > 0 {
		for _, v := range matches {
			bundleFile, err := os.Open(filepath.Join(bundleBasePath, v[1]))
			if err != nil {
				return nil, err
			}

			contentType, err := GetContentType(filepath.Ext(bundleFile.Name()))
			if err != nil {
				return bundleComponents, err
			}

			component := MultipartUploadComponent{
				Data:        bundleFile,
				Name:        strings.Split(filepath.Base(bundleFile.Name()), ".")[0],
				FileName:    filepath.Base(bundleFile.Name()),
				ContentType: contentType,
			}
			bundleComponents = append(bundleComponents, component)

			if contentType == ContentTypeHtml {
				additionalWebComponents, err := getAdditionalWebComponents(bundleBasePath)
				if err != nil {
					return nil, err
				}
				bundleComponents = append(bundleComponents, additionalWebComponents...)
			}
		}
	}

	return bundleComponents, nil
}

func getAdditionalWebComponents(bundleBasePath string) ([]MultipartUploadComponent, error) {
	var components []MultipartUploadComponent
	files, err := ioutil.ReadDir(bundleBasePath)

	if err != nil {
		return nil, err
	}
	for _, f := range files {
		extension := strings.Split(f.Name(), ".")[1]
		if extension == "css" || extension == "js" || extension == "manifest" {
			contentType, err := GetContentType("." + extension)
			if err != nil {
				return nil, err
			}
			bundleFile, err := os.Open(filepath.Join(bundleBasePath, f.Name()))
			if err != nil {
				return nil, err
			}
			components = append(components, MultipartUploadComponent{
				Data:        bundleFile,
				Name:        strings.Split(filepath.Base(bundleFile.Name()), ".")[0],
				FileName:    filepath.Base(bundleFile.Name()),
				ContentType: contentType,
			})
		}
	}
	return components, nil
}

func (c *Client) CreateArticle(article io.Reader, bundleComponents []MultipartUploadComponent, metadata *Metadata) (*ReadArticleResponse, error) {
	url := fmt.Sprintf("%s/channels/%s/articles", c.BaseURL, c.ChannelID)

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	multipartComponents := []MultipartUploadComponent{
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

	multipartComponents = append(multipartComponents, bundleComponents...)

	req, err := c.prepareMultipartRequest(
		multipartComponents,
		url,
	)

	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.Errorf("%s returned a %d . reason: %s", url, resp.StatusCode, string(body))
	}

	var readArticleResp ReadArticleResponse
	err = json.Unmarshal(body, &readArticleResp)
	if err != nil {
		return nil, err
	}

	return &readArticleResp, resp.Body.Close()
}

func (c *Client) UpdateArticle(articleId string, revision string, article io.Reader, bundleComponents []MultipartUploadComponent, metadata *Metadata) (*ReadArticleResponse, error) {
	url := fmt.Sprintf("%s/articles/%s", c.BaseURL, articleId)

	metadata.Data.Revision = revision

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	metadata.Data.Revision = revision

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

	parts = append(parts, bundleComponents...)

	req, err := c.prepareMultipartRequest(
		parts,
		url,
	)

	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("%s returned a %d . reason: %s", url, resp.StatusCode, string(body))
	}

	var readArticleResp ReadArticleResponse
	err = json.Unmarshal(body, &readArticleResp)
	if err != nil {
		return nil, err
	}

	return &readArticleResp, resp.Body.Close()
}

func (c *Client) UpdateArticleMetadata(articleId string, metadata *Metadata) (*ReadArticleResponse, error) {
	url := fmt.Sprintf("%s/articles/%s", c.BaseURL, articleId)

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("%s returned a %d . reason: %s", url, resp.StatusCode, string(body))
	}

	var readArticleResp ReadArticleResponse
	err = json.Unmarshal(body, &readArticleResp)
	if err != nil {
		return nil, err
	}

	return &readArticleResp, resp.Body.Close()
}

func (c *Client) PromoteArticles(sectionId string, articleIds []string) (*PromoteArticlesResponse, error) {
	url := fmt.Sprintf("%s/sections/%s/promotedArticles", c.BaseURL, sectionId)

	promotedArticles := PromoteArticlesRequest{}
	if len(articleIds) == 0 {
		articleIds = make([]string, 0)
	}
	promotedArticles.Data.PromotedArticles = articleIds

	bodyBytes, err := json.Marshal(promotedArticles)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))

	auth, err := c.getAuthorization(http.MethodPost, url, "", ioutil.NopCloser(bytes.NewReader(bodyBytes)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	b, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("%s returned a %d . reason: %s", url, resp.StatusCode, string(b))
	}

	var promoteArticlesResponse PromoteArticlesResponse
	err = json.Unmarshal(b, &promoteArticlesResponse)
	if err != nil {
		return nil, err
	}

	return &promoteArticlesResponse, resp.Body.Close()
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

	if resp.StatusCode != 204 {
		b, _ := ioutil.ReadAll(resp.Body)
		return errors.Errorf("%s returned a %d . reason: %s", url, resp.StatusCode, string(b))
	}
	return nil
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
		partBytes, err := ioutil.ReadAll(v.Data)
		if err != nil {
			return nil, err
		}
		h.Set("Content-Length", fmt.Sprintf("%d", len(partBytes)))
		part, err := writer.CreatePart(h)
		if err != nil {
			return nil, err
		}
		_, err = part.Write(partBytes)
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

func GetContentType(extension string) (ContentType, error) {
	switch strings.ToLower(extension) {
	case ".jpg", ".jpeg":
		return ContentTypeJpeg, nil
	case ".png":
		return ContentTypePng, nil
	case ".gif":
		return ContentTypeGif, nil
	case ".html":
		return ContentTypeHtml, nil
	case ".manifest":
		return ContentTypeManifest, nil
	case ".css":
		return ContentTypeCss, nil
	case ".js":
		return ContentTypeJs, nil
	default:
		return "", errors.New(fmt.Sprintf("Could not match extension %s to a valid content type", extension))
	}
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}
