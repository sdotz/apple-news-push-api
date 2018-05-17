package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"time"
)

//Builds the Authorization header according to the spec defined here: https://developer.apple.com/library/content/documentation/General/Conceptual/News_API_Ref/Security.html#//apple_ref/doc/uid/TP40015409-CH5-SW1
func (c *Client) getAuthorization(httpMethod string, url string, contentType string, body io.ReadCloser) (string, error) {
	defer body.Close()
	timeNow := time.Now().UTC().Format(time.RFC3339)
	apiSecretDecoded, err := base64.StdEncoding.DecodeString(c.APISecret)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, apiSecretDecoded)

	//The beginning of the "canonical request".The body will then be appended onto it.
	_, err = mac.Write([]byte(fmt.Sprintf("%s%s%s%s", httpMethod, url, timeNow, contentType)))
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(mac, body); err != nil {
		return "", err
	}

	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("HHMAC; key=%s; signature=%s; date=%s", c.APIKey, signature, timeNow), nil
}
