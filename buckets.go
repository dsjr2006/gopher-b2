package gopherb2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/uber-go/zap"
)

type Buckets struct {
	Bucket []Bucket `json:"buckets"`
}
type Bucket struct {
	AccountID      string   `json:"accountId"`
	BucketID       string   `json:"bucketId"`
	BucketName     string   `json:"bucketName"`
	BucketType     string   `json:"bucketType"`
	LifecycleRules []string `json:"lifecycleRules"`
	Revision       int      `json:"revision"`
}

// Creates new B2 bucket and returns API response
func B2CreateBucket(bucketName string, bucketPublic bool) {
	//TODO: Check bucket name validity

	if len(bucketName) < 6 {
		logger.Fatal("Bucket Name must be at least 6 chars",
			zap.String("Bucket Name too short", bucketName),
		)
	}

	// Public or private bucketName
	var bucketType = "allPrivate"
	if bucketPublic == true {
		bucketType = "allPublic"
	}

	// Authorize and Get API Token
	authorizationResponse := AuthorizeAcct()

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
	if apiResponse.Status == "200 OK" {
		logger.Debug("Create New Bucket Successful",
			zap.String("Bucket Name:", bucketName),
		)
	} else {
		logger.Panic("Could not create new Bucket",
			zap.String("API Resp Body:", string(apiResponse.Body)),
		)
	}
	// Parse JSON 'Bucket' Response
	var bucket Bucket
	err = json.Unmarshal(apiResponse.Body, &bucket)
	if err != nil {
		fmt.Println("Bucket JSON Parse Failed", err)
	}
	logger.Info("New Bucket Created",
		zap.String("Bucket Name:", bucketName),
		zap.String("Bucket ID:", bucket.BucketID),
	)

	return
}

// B2GetBuckets calls authorizeAccount then connects to API to request list of all B2 buckets and information, returns type 'Buckets' and error
func GetBuckets() (Buckets, error) {
	// Authorize and Get API Token
	authorizationResponse := AuthorizeAcct()

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
		logger.Warn("List Buckets Failed.",
			zap.Error(err),
		)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	var apiResponse Response
	apiResponse = Response{Header: resp.Header, Status: resp.Status, Body: respBody}
	if resp.Status != "200 OK" {
		logger.Fatal("Bucket List Request API Error",
			zap.String("API Resp Status", resp.Status),
			zap.String("API Resp Body", string(respBody)),
		)
	}
	// Parse JSON 'Bucket' Response
	var buckets Buckets
	err = json.Unmarshal(apiResponse.Body, &buckets)
	if err != nil {
		logger.Fatal("Bucket JSON Parse Failed",
			zap.Error(err),
		)
	}

	return buckets, err
}

// PrintBuckets Diplays list of buckets in console
func PrintBuckets(buckets Buckets) error {
	if buckets.Bucket != nil {
		writer := new(tabwriter.Writer)
		fmt.Println("B2 Buckets")
		// Format to '|' separated columns with no min width and blank padding char
		writer.Init(os.Stdout, 0, 5, 1, ' ', 0)
		fmt.Fprintln(writer, "-ID-\t -NAME-\t -TYPE-")
		for i := 0; i < len(buckets.Bucket); i++ {
			fmt.Fprintln(writer, buckets.Bucket[i].BucketID+"\t", buckets.Bucket[i].BucketName+"\t", buckets.Bucket[i].BucketType+"\t")
		}
		fmt.Fprintln(writer)
		writer.Flush()
		return nil
	}
	return errors.New("No Buckets to print")
}
