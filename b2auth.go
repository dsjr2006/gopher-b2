package main

import (
	"encoding/base64"
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
func main () {
	//readConfiguration()
	fmt.Println(string(authorizeAccount()))
}
// Calling this function reads settings.toml file in "/config" and returns the HTTP response body
func authorizeAccount() []byte {
	var Config Configuration
	viper.SetConfigName("settings")     // no need to include file extension
  viper.AddConfigPath("config")  // set the path of your config file
	err := viper.ReadInConfig()
  if err != nil {
    fmt.Println("Config file not found...")
  } else {
		Config.ACCOUNT_ID = viper.GetString("Account1.ACCOUNT_ID")
		Config.APPLICATION_ID = viper.GetString("Account1.APPLICATION_ID")
		Config.API_URL = viper.GetString("Account1.API_URL")
	}
	// Encode credentials to base64
	credentials := base64.StdEncoding.EncodeToString([]byte(Config.ACCOUNT_ID + ":" + Config.APPLICATION_ID))

	// Request (POST https://api.backblazeb2.com/b2api/v1/b2_authorize_account)
	json := []byte(`{}`)
	body := bytes.NewBuffer(json)

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

	// Display Results
	/*
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	fmt.Println("response Body : ", string(respBody))
	*/

	return respBody
}
