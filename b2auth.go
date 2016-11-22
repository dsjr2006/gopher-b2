// B2 API Authorization Related Functions
package github.com/dsjr2006/gopher-b2/b2auth

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	ACCOUNT_ID     = "0e1000668b00"
	APPLICATION_ID = "001t38150d55555estc00ba25f666666d380c23b9"
	API_URL        = "https://api.backblazeb2.com/b2api/v1/b2_authorize_account"
)

func sendRequest() {
	// Encode credentials to base64
	credentials := base64.StdEncoding.EncodeToString([]byte(ACCOUNT_ID + ":" + APPLICATION_ID))
	fmt.Println(credentials)

	// Request (POST https://api.backblazeb2.com/b2api/v1/b2_authorize_account)
	json := []byte(`{}`)
	body := bytes.NewBuffer(json)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", API_URL, body)

	// Headers
	req.Header.Add("Authorization", "Basic "+credentials)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Display Results
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	fmt.Println("response Body : ", string(respBody))
}
