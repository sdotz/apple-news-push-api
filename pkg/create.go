package pkg

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
	MaturityRating                 `json:"maturityRating,omitempty"`
	IsCandidateToBeFeatured bool   `json:"isCandidateToBeFeatured,omitempty"`
	IsHidden                bool   `json:"isHidden,omitempty"`
}

func CreateArticle(channelId string, article io.Reader, metadata *Metadata, apiKey string, apiSecret string, baseUrl string) {

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

func prepareMultipartRequest(parts []MultipartUploadComponent, url string, apiKey string, apiSecret string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, v := range parts {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				escapeQuotes(v.Name), escapeQuotes(v.FileName)))
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
