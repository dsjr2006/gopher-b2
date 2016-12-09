package gopherb2

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Configuration struct {
	ACCOUNT_ID     string
	APPLICATION_ID string
	API_URL        string
}
type Response struct {
	Header http.Header
	Status string
	Body   []byte
}
type APIAuthorization struct {
	AccountID          string `json:"accountId"`
	ApiURL             string `json:"apiUrl"`
	AuthorizationToken string `json:"authorizationToken"`
	DownloadURL        string `json:"downloadURL"`
	MinimumPartSize    int    `json:"minimumPartSize"`
}
type Buckets struct {
	Bucket []struct {
		AccountID      string   `json:"accountId"`
		BucketID       string   `json:"bucketId"`
		BucketName     string   `json:"bucketName"`
		BucketType     string   `json:"bucketType"`
		LifecycleRules []string `json:"lifecycleRules"`
		Revision       int      `json:"revision"`
	} `json:"buckets"`
}
type UploadURL struct {
	AuthorizationToken string `json:"authorizationToken"`
	BucketId           string `json:"bucketId"`
	URL                string `json:"uploadUrl"`
}
// Setup Logging
var logger = zap.New(
	zap.NewJSONEncoder(),
)
// Simple Error check for fatal errors
func checkError(e error) {
  if e != nil {
		logger.Fatal("checkError failed",
			zap.Error(e),
		)
		panic(e)
  }
}
// Calling this function reads settings.toml file in "/config" , calls B2 API , then returns the response as APIAuthorization struct
func B2AuthorizeAccount() APIAuthorization {
	var Config Configuration
	viper.SetConfigName("settings") // no need to include file extension
	viper.AddConfigPath("config")   // set the path of your config file
	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatal("No Configuration file found, Cannot Attempt Authorization with API.")
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
		logger.Warn("API Auth Request Failed.",
			zap.Error(err),
		)
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
	if err != nil {
		fmt.Println("API Auth JSON Parse Failed", err)
		logger.Fatal("Cannot parse API Auth Response JSON.",
			zap.Error(err),
		)
	}

	return apiAuth
}

// Calls authorizeAccount then connects to API to request list of all B2 buckets and information, returns type 'Buckets'
func B2ListBuckets() Buckets {
	// Authorize and Get API Token
	authorizationResponse := B2AuthorizeAccount()

	// Request (POST https://api001.backblazeb2.com/b2api/v1/b2_list_buckets)

	jsonData := []byte(`{"accountId": "` + authorizationResponse.AccountID + `"}`)
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
	defer resp.Body.Close()

	var apiResponse Response
	apiResponse = Response{Header: resp.Header, Status: resp.Status, Body: respBody}
	// Display Results
	/*
		fmt.Println("response Status : ", resp.Status)
		fmt.Println("response Headers : ", resp.Header)
		fmt.Println("response Body : ", string(respBody))
	*/

	// Parse JSON 'Bucket' Response
	var buckets Buckets
	err = json.Unmarshal(apiResponse.Body, &buckets)
	if err != nil {
		fmt.Println("Bucket JSON Parse Failed", err)
	}
	/*
		fmt.Println("Bucket 0: " + string(bucketList.Buckets[0]))
		fmt.Printf("Buckets: %+v\n", buckets)
		fmt.Println("Bucket 0 Name: " + (buckets.Bucket[0].BucketName))
	*/
	return buckets
}

// Creates new B2 bucket and returns API response
func B2CreateBucket(bucketName string, bucketPublic bool) Response {
	// Check bucket name validity

	// Public or private bucketName
	var bucketType = "allPrivate"
	if bucketPublic == true {
		bucketType = "allPublic"
	}

	// Authorize and Get API Token
	authorizationResponse := B2AuthorizeAccount()

	// Request (POST https://api001.backblazeb2.com/b2api/v1/b2_create_bucket)

	jsonData := []byte(`{"accountId": "` + authorizationResponse.AccountID + `", "bucketName":"` + bucketName + `", "bucketType":"` + bucketType + `" }`)
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
	defer resp.Body.Close()

	var apiResponse Response
	apiResponse = Response{Header: resp.Header, Status: resp.Status, Body: respBody}

	return apiResponse
}

// Requests Upload URL from API and returns 'UploadURL'
func B2GetUploadURL(bucketId string) UploadURL {
	// Authorize and Get API Token
	authorizationResponse := B2AuthorizeAccount()

	// Get Upload URL (POST https://api001.backblazeb2.com/b2api/v1/b2_get_upload_url)

	jsonData := []byte(`{"bucketId": "` + bucketId + `"}`)
	body := bytes.NewBuffer(jsonData)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", authorizationResponse.ApiURL+"/b2api/v1/b2_get_upload_url", body)

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
	defer resp.Body.Close()
	var apiResponse Response
	apiResponse = Response{Header: resp.Header, Status: resp.Status, Body: respBody}

	var uploadURL UploadURL

	err = json.Unmarshal(apiResponse.Body, &uploadURL)
	if err != nil {
		fmt.Println("Bucket JSON Parse Failed", err)
		log.Fatal(err)
	}

	return uploadURL
}

// Upload single file to B2
func B2UploadFile(bucketId string, filePath string) Response {
	// Authorize and Get Upload URL
	uploadURL := B2GetUploadURL(bucketId)

	fmt.Println("File: " + filePath + "\n")
	file, err := os.Open(filePath)
	defer file.Close()
	checkError(err)
	fileInfo, err := file.Stat()
	checkError(err)

	// Create client
	client := &http.Client{}
	// Request Body
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		log.Fatal(err)
	}
	checkError(err)
	body := buffer
	// Get File Modification Time as int64 value in milliseconds since midnight, January 1, 1970 UTC
	fileModTimeMillis := fileInfo.ModTime().UnixNano() / 1000000

	// Create request
	req, err := http.NewRequest("POST", uploadURL.URL, body)

	// Headers
	req.Header.Add("Authorization", uploadURL.AuthorizationToken)
	req.Header.Add("Content-Type", "b2/x-auto")
	req.Header.Add("Content-Length", string(fileInfo.Size()))
	req.Header.Add("X-Bz-Content-Sha1", fileSHA1(filePath))
	req.Header.Add("X-Bz-File-Name", fileInfo.Name()) //Need to encode names properly! according to B2 docs
	req.Header.Add("X-Bz-Info-src_last_modified_millis", fmt.Sprintf("%d", fileModTimeMillis))
	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		logger.Warn("Upload Request Failed",
			zap.Error(err),
		)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	var apiResponse Response
	apiResponse = Response{Header: resp.Header, Status: resp.Status, Body: respBody}

	return apiResponse
}
// Begin Large File Upload
func B2StartLargeFile(bucketId string, filePath string) Response {
	// Authorize
	apiAuth := B2AuthorizeAccount()

	// Open File and Get File Stats
	file, err := os.Open(filePath)
	defer file.Close()
	checkError(err)
	fileInfo, err := file.Stat()
	checkError(err)
	// Get File Modification Time as int64 value in milliseconds since midnight, January 1, 1970 UTC
	fileModTimeMillis := fileInfo.ModTime().UnixNano() / 1000000
	// Get File sha1
	largeFileSHA1 := fileSHA1(filePath)
	if largeFileSHA1 == "fail" {
		logger.Fatal("Cannot parse API Auth Response JSON.",)
	}

	// Create client
	client := &http.Client{}
	// Request Body : JSON object
	jsonBody := []byte(`{"fileInfo": {"large_file_sha1": "`+ largeFileSHA1 +`","src_last_modified_millis": "`+ fmt.Sprintf("%d", fileModTimeMillis) +`"},"bucketId": "`+ bucketId +`","fileName": "`+fileInfo.Name()+`","contentType": "b2/x-auto"}`)
	body := bytes.NewBuffer(jsonBody)
	fmt.Println(body)

	// Create request
	req, err := http.NewRequest("POST", "https://api001.backblazeb2.com/b2api/v1/b2_start_large_file", body)

	// Headers
	req.Header.Add("Authorization", apiAuth.AuthorizationToken)

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	var apiResponse Response
	apiResponse = Response{Header: resp.Header, Status: resp.Status, Body: respBody}

	return apiResponse
}
func fileSHA1(filePath string) string {
	file, err := os.Open(filePath)
	defer file.Close()
	checkError(err)
	hash := sha1.New()
	// Copy file into hash interface
	if _, err := io.Copy(hash, file); err != nil {
		logger.Warn("File SHA1 Hash Failure",
			zap.Error(err),
		)
		return "fail"
	}
	// Get 20 bytes hash
	hashAsBytes := hash.Sum(nil)[:20]

	return hex.EncodeToString(hashAsBytes)
}
func encodeFilename(filePath string) string {
	file, err := os.Open(filePath)
	defer file.Close()
	checkError(err)
	fileInfo, err := file.Stat()
	checkError(err)
	encodedFilename := string(fileInfo.Name())
	fmt.Println(encodedFilename)
	return encodedFilename
}
