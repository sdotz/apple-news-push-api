package pkg

import (
	"time"
	"fmt"
	"crypto/sha256"
	"crypto/hmac"
	"encoding/base64"
	"log"
)

func getAuthorization(httpMethod string, url string, apiKey string, apiSecret string, contentType string, body string) string {
	timeNow := time.Now().UTC().Format(time.RFC3339)
	canonicalRequest := fmt.Sprintf("%s%s%s%s%s", httpMethod, url, timeNow, contentType, body)
	apiSecretDecoded, err := base64.StdEncoding.DecodeString(apiSecret)
	if err != nil {
		log.Fatal(err)
	}
	mac := hmac.New(sha256.New, apiSecretDecoded)

	mac.Write([]byte(canonicalRequest))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("HHMAC; key=%s; signature=%s; date=%s", apiKey, signature, timeNow)
}
