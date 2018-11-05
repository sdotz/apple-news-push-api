package api

import (
	"encoding/json"
	"bytes"
	"io/ioutil"
	"net/http"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

const (
	CountryEU = "EU"
	CountryGB = "GB"
	CountryUS = "US"
)

type NotificationData struct {
	AlertBody string   `json:"alertBody"`
	Countries []string `json:"countries,omitempty"`
}

type NotificationRequest struct {
	NotificationData `json:"data"`
}

type NotificationResponse struct {
	Data struct {
		CreatedAt  time.Time `json:"createdAt"`
		ModifiedAt time.Time `json:"modifiedAt"`
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Links struct {
			Article string `json:"article"`
		} `json:"links"`
		AlertBody string   `json:"alertBody"`
		Countries []string `json:"countries"`
	} `json:"data"`
	Meta struct {
		Quotas struct {
			Daily struct {
				Sent  int `json:"sent"`
				Limit int `json:"limit"`
			} `json:"daily"`
		} `json:"quotas"`
	} `json:"meta"`
}

func (c *Client) SendNotification(articleId string, alertBody string, countries []string, ignoreWarnings bool) (*NotificationResponse, error) {
	if !ignoreWarnings {
		err := validateAlertBodyLength(alertBody)
		if err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("%s/articles/%s/notifications", c.BaseURL, articleId)

	notificationReq := NotificationRequest{
		NotificationData: NotificationData{
			AlertBody: alertBody,
			Countries: countries,
		},
	}

	bodyJsonBytes, err := json.Marshal(notificationReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyJsonBytes))

	if err != nil {
		return nil, err
	}

	auth, err := c.getAuthorization(http.MethodPost, url, "application/json", ioutil.NopCloser(bytes.NewReader(bodyJsonBytes)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	b, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return nil, errors.Errorf("%s returned a %d . reason: %s", url, resp.StatusCode, string(b))
	}

	var notificationResponse NotificationResponse

	if err := json.Unmarshal(b, &notificationResponse); err != nil {
		return nil, err
	}

	return &notificationResponse, nil
}

func validateAlertBodyLength(alertBody string) error {
	if len(alertBody) > 500 {
		return errors.New(fmt.Sprintf("Warning: Alert was longer than max length: %d/500 chars. The rest would be truncated", len(alertBody)))
	}
	if len(alertBody) > 130 {
		return errors.New(fmt.Sprintf("Warning: Alert was longer than recommedned length: %d/130 chars", len(alertBody)))
	}
	return nil
}
