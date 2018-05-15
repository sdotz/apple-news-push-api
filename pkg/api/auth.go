package api

import (
	"time"
	"fmt"
	"crypto/sha256"
	"crypto/hmac"
	"encoding/base64"
	"log"
	"io"
)

//Builds the Authorization header according to the spec defined here: https://developer.apple.com/library/content/documentation/General/Conceptual/News_API_Ref/Security.html#//apple_ref/doc/uid/TP40015409-CH5-SW1
func getAuthorization(httpMethod string, url string, apiKey string, apiSecret string, contentType string, body io.ReadCloser) string {
	defer body.Close()
	timeNow := time.Now().UTC().Format(time.RFC3339)
	apiSecretDecoded, err := base64.StdEncoding.DecodeString(apiSecret)
	if err != nil {
		log.Fatal(err)
	}
	mac := hmac.New(sha256.New, apiSecretDecoded)

	//The beginning of the "canonical request".The body will then be appended onto it.
	_, err = mac.Write([]byte(fmt.Sprintf("%s%s%s%s", httpMethod, url, timeNow, contentType)))
	if err != nil {
		log.Fatal(err)
	}

	if _, err := io.Copy(mac, body); err != nil {
		log.Fatal(err)
	}

	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("HHMAC; key=%s; signature=%s; date=%s", apiKey, signature, timeNow)
}
