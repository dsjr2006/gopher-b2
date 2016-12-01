package gopherb2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
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
type BucketList struct {
	Buckets []json.RawMessage `json:"buckets"`
}
type Buckets struct {
	Bucket []struct {
		AccountID string `json:"accountId"`
		BucketID string `json:"bucketId"`
		BucketName string `json:"bucketName"`
		BucketType string `json:"bucketType"`
		LifecycleRules []string `json:"lifecycleRules"`
		Revision int `json:"revision"`
	} `json:"buckets"`
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
		Config.API_URL = viper.GetString("Account1.API_URL")
	}
	// Encode credentials to base64
	credentials := base64.StdEncoding.EncodeToString([]byte(Config.ACCOUNT_ID + ":" + Config.APPLICATION_ID))

	// Request (POST https://api.backblazeb2.com/b2api/v1/b2_authorize_account)
	jsonData := []byte(`{}`)
	body := bytes.NewBuffer(jsonData)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", Config.API_URL+"b2_authorize_account", body)

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
		fmt.Println("API Auth JSON Parse Failed", err)
	}

	return apiAuth
}
// Calls authorizeAccount then connects to API to request list of all B2 buckets and returns Response type
func listBuckets() Response {
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

	var apiResponse Response
	apiResponse = Response{Header: resp.Header, Status: resp.Status, Body:respBody}
	// Display Results
	/*
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	fmt.Println("response Body : ", string(respBody))
	*/
	bucketList := &BucketList{}
	var buckets Buckets
	err = json.Unmarshal(respBody, &bucketList)
	if(err != nil){
		fmt.Println("Buckets List JSON Parse Failed", err)
	}

	err = json.Unmarshal(respBody, &buckets)
	if(err != nil){
		fmt.Println("Bucket JSON Parse Failed", err)
		log.Fatal(err)
	}

	fmt.Println("Bucket 0: " + string(bucketList.Buckets[0]))
	fmt.Printf("Buckets: %+v\n", buckets)
	fmt.Println("Bucket 0 Name: " + (buckets.Bucket[0].BucketName))

	return apiResponse
}
// Creates new B2 bucket and returns API response
func createBucket(bucketName string, bucketPublic bool) Response {
	// Check bucket name validity

	// Public or private bucketName
	var bucketType = "allPrivate"
	if bucketPublic == true {
		bucketType = "allPublic"
	}


	// Authorize and Get API Token
	authorizationResponse:= authorizeAccount()

	// Request (POST https://api001.backblazeb2.com/b2api/v1/b2_create_bucket)

	jsonData := []byte(`{"accountId": "`+ authorizationResponse.AccountID +`", "bucketName":"`+ bucketName +`", "bucketType":"`+ bucketType +`" }`)
	body := bytes.NewBuffer(jsonData)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", authorizationResponse.ApiURL+"/b2api/v1/b2_create_bucket", body)

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

	var apiResponse Response
	apiResponse = Response{Header: resp.Header, Status: resp.Status, Body:respBody}

	return apiResponse
}
