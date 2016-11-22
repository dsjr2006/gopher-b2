package gopherb2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"bytes"
	"github.com/spf13/viper"
)

type Configuration struct {
  ACCOUNT_ID	string
  APPLICATION_ID	string
  API_URL	string
}
type Response struct {
	Header http.Header
	Status string
	Body []byte
}
type APIAuthorization struct {
	AccountID string `json:"accountId"`
	ApiURL string `json:"apiUrl"`
	AuthorizationToken string `json:"authorizationToken"`
	DownloadURL string `json:"downloadURL"`
	MinimumPartSize int `json:"minimumPartSize"`
}
func main () {
	//readConfiguration()
	//fmt.Println(string(authorizeAccount()))
}

// Calling this function reads settings.toml file in "/config" , calls B2 API , then returns the response as APIAuthorization struct
func authorizeAccount() APIAuthorization  {
	var Config Configuration
	viper.SetConfigName("settings")     // no need to include file extension
  viper.AddConfigPath("config")  // set the path of your config file
	err := viper.ReadInConfig()
  if err != nil {
    fmt.Println("Config file not found...")
  } else {
		Config.ACCOUNT_ID = viper.GetString("Account1.ACCOUNT_ID")
		Config.APPLICATION_ID = viper.GetString("Account1.APPLICATION_ID")
		Config.API_URL = viper.GetString("Account1.API_AuthorizationURL")
	}
	// Encode credentials to base64
	credentials := base64.StdEncoding.EncodeToString([]byte(Config.ACCOUNT_ID + ":" + Config.APPLICATION_ID))

	// Request (POST https://api.backblazeb2.com/b2api/v1/b2_authorize_account)
	jsonData := []byte(`{}`)
	body := bytes.NewBuffer(jsonData)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", Config.API_URL, body)

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
	defer resp.Body.Close()

	//apiResponse := Response{Header: resp.Header, Status: resp.Status, Body: respBody}
	// Display Results
	/*
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	fmt.Println("response Body : ", string(respBody))
	*/
	var apiAuth APIAuthorization
	err = json.Unmarshal(respBody, &apiAuth)
	if(err != nil){
		fmt.Println("JSON Parse Failed", err)
	}

	return apiAuth
}
func listBuckets() []byte {
	// Authorize and Get API Token
	authorizationResponse:= authorizeAccount()

	// Request (POST https://api001.backblazeb2.com/b2api/v1/b2_list_buckets)

	jsonData := []byte(`{"accountId": "`+ authorizationResponse.AccountID +`"}`)
	body := bytes.NewBuffer(jsonData)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", authorizationResponse.ApiURL+"/b2api/v1/b2_list_buckets", body)

	// Headers
	req.Header.Add("Authorization", authorizationResponse.AuthorizationToken)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Display Results
	/*
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	fmt.Println("response Body : ", string(respBody))
	*/

	return respBody
}
