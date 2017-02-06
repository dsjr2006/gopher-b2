package gopherb2

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// Log
// Log Level
var logDebug = false
var logger = zap.New(
	zap.NewJSONEncoder(),
	zap.DebugLevel,
)

type APIAuthorization struct {
	AccountID          string `json:"accountId"`
	ApiURL             string `json:"apiUrl"`
	AuthorizationToken string `json:"authorizationToken"`
	DownloadURL        string `json:"downloadURL"`
	MinimumPartSize    int    `json:"minimumPartSize"`
}

// Calling this function reads settings.toml file in "/config" , calls B2 API , then returns the response as APIAuthorization struct
func B2AuthorizeAccount() APIAuthorization {
	var Config Configuration
	viper.SetConfigName("settings")  // no need to include file extension
	viper.AddConfigPath("../config") // set the path of your config file
	err := viper.ReadInConfig()
	viper.AddConfigPath("config") // set the path of your config file
	err = viper.ReadInConfig()
	if err != nil {
		logger.Fatal("No Configuration file found, Cannot Attempt Authorization with API.")
	} else {
		Config.ACCOUNT_ID = viper.GetString("Account1.ACCOUNT_ID")
		logger.Debug("Obtained Account ID from Configuration file",
			zap.String("Account ID:", Config.ACCOUNT_ID),
		)
		Config.APPLICATION_ID = viper.GetString("Account1.APPLICATION_ID")
		logger.Debug("Obtained Account ID from Configuration file",
			zap.String("Application ID:", Config.APPLICATION_ID),
		)
		Config.API_URL = viper.GetString("Account1.API_URL")
		logger.Debug("Obtained Account ID from Configuration file",
			zap.String("API URL:", Config.API_URL),
		)
	}
	if Config.ACCOUNT_ID == "000" {
		logger.Fatal("Account ID set to default. Update with your Account Id from Backblaze Settings.")
	} else if Config.APPLICATION_ID == "000" {
		logger.Fatal("Application ID set to default. Update with your Application Id from Backblaze Settings.")
	}
	// Encode credentials to base64
	credentials := base64.StdEncoding.EncodeToString([]byte(Config.ACCOUNT_ID + ":" + Config.APPLICATION_ID))

	// Request (POST https://api.backblazeb2.com/b2api/v1/b2_authorize_account)
	jsonData := []byte(`{}`)
	body := bytes.NewBuffer(jsonData)
	logger.Debug("Preparing to send API Auth Request",
		zap.String("Credentials Encoded:", credentials),
	)

	// Create client
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("POST", Config.API_URL+"b2_authorize_account", body)
	if err != nil {
		logger.Fatal("Creating API Auth Request Failed.",
			zap.Error(err),
		)
	}

	// Headers
	req.Header.Add("Authorization", "Basic "+credentials)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	// Troubleshooting
	fmt.Println("Auth Request")
	for k, v := range req.Header {
		fmt.Printf("\n%v: %v", k, v)
	}
	fmt.Printf("\n%v", string(jsonData))

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		logger.Fatal("API Auth Request Failed.",
			zap.Error(err),
		)
	}
	logger.Debug("Received API Authorization response.")

	var apiAuth APIAuthorization

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	err = json.Unmarshal(respBody, &apiAuth)

	//Troubleshooting
	fmt.Println("Auth Response")
	for k, v := range resp.Header {
		fmt.Printf("\n%v: %v", k, v)
	}
	fmt.Printf("\n%v", string(respBody))

	if err != nil {
		fmt.Println("API Auth JSON Parse Failed", err)
		logger.Fatal("Cannot parse API Auth Response JSON.",
			zap.Error(err),
		)
	}

	if resp.Status != "200 OK" {
		logger.Fatal("Authorization with Backblaze B2 API Failed",
			zap.String("API Resp Body:", string(respBody)),
		)
	}

	// Check API Response matches config
	if apiAuth.AccountID != Config.ACCOUNT_ID {
		logger.Fatal("API Account ID Response does not match Account ID in Config.",
			zap.String("API Resp Acct ID", apiAuth.AccountID),
			zap.String("Config Acct ID", Config.ACCOUNT_ID),
		)
	}

	return apiAuth
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

	// Troubleshooting
	fmt.Println("Upload URL Request")
	for k, v := range req.Header {
		fmt.Printf("\n%v: %v", k, v)
	}
	fmt.Printf("\n%v", string(jsonData))

	logger.Debug("Preparing to send Get Upload URL request.")

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	//Troubleshooting
	fmt.Println("Upload URL Response")
	for k, v := range resp.Header {
		fmt.Printf("\n%v: %v", k, v)
	}
	fmt.Printf("\n%v", string(respBody))

	var apiResponse Response
	apiResponse = Response{Header: resp.Header, Status: resp.Status, Body: respBody}

	var uploadURL UploadURL

	err = json.Unmarshal(apiResponse.Body, &uploadURL)
	if err != nil {
		logger.Fatal("Bucket JSON Parse Failed",
			zap.Error(err),
		)
	}
	logger.Debug("Get Upload URL response received from API")

	return uploadURL
}

func B2FinishLargeFile(largeFile LargeFile) error {
	apiAuth := B2AuthorizeAccount()

	// Create SHA1 array of completed files
	var partSha1Array string
	var buffer bytes.Buffer
	for i := 0; i < len(largeFile.Temp); i++ {
		buffer.WriteString("\u0022") // Add double quotation mark --> "
		buffer.WriteString(largeFile.Temp[i].SHA1)
		buffer.WriteString("\u0022")
		if i != len(largeFile.Temp)-1 {
			buffer.WriteString(",") // Only add trailing comma if item is not last
		}
	}
	partSha1Array = buffer.String()

	// Create client
	client := &http.Client{}
	// Request Body : JSON object with fileID & array of SHA1 hashes of files transmitted
	jsonBody := []byte(`{"fileId": "` + largeFile.FileID + `", "partSha1Array":[` + partSha1Array + `]}`)
	body := bytes.NewBuffer(jsonBody)

	// Create request
	req, err := http.NewRequest("POST", apiAuth.ApiURL+"/b2api/v1/b2_finish_large_file", body)

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

	if resp.Status != "200 OK" {
		logger.Warn("Finish B2 Large File Failed",
			zap.String("HTTP Status Response", resp.Status),
			zap.String("HTTP Resp Body", string(respBody)),
		)
	}
	if resp.Status == "200 OK" {
		logger.Info("Finish Large File Upload Completed",
			zap.String("Filepath", largeFile.OrigPath),
			zap.String("B2 File ID", largeFile.FileID),
		)
	}

	return err
}

func B2GetUploadPartURL(fileId string) UploadPartResponse {
	apiAuth := B2AuthorizeAccount()

	// Create client
	client := &http.Client{}
	// Request Body : JSON object with fileId
	jsonBody := []byte(`{"fileId": "` + fileId + `"}`)
	body := bytes.NewBuffer(jsonBody)

	// Create request
	req, err := http.NewRequest("POST", apiAuth.ApiURL+"/b2api/v1/b2_get_upload_part_url", body)

	// Headers
	req.Header.Add("Authorization", apiAuth.AuthorizationToken)

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		logger.Fatal("Error requesting Part Upload URL",
			zap.Error(err),
		)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	var uploadPartResponse UploadPartResponse
	if resp.Status != "200 OK" {
		logger.Fatal("Could not obtain Part Upload URL", zap.String("Response", string(respBody)))
	} else if resp.Status == "200 OK" {
		err = json.Unmarshal(respBody, &uploadPartResponse)
		if err != nil {
			logger.Fatal("Upload Part Response JSON Parse Failed", zap.Error(err))
		}
		logger.Info("Obtained Upload Part URL",
			zap.String("B2 File ID", uploadPartResponse.FileID),
		)
	}
	return uploadPartResponse
}
