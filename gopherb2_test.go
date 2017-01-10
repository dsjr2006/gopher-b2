package gopherb2

import (
	"fmt"
	"testing"
)

// Test authorizeAccount
func TestToReturnAuthorization(t *testing.T) {
	apiResponse := B2AuthorizeAccount()
	fmt.Println("\nAccount ID: " + apiResponse.AccountID)
	fmt.Println("\n^ Authorization Test Completed\n")
}

// Test listBuckets
func TestToReturnBucketList(t *testing.T) {
	allBuckets := B2ListBuckets()
	//fmt.Println("\nBuckets Response Body:\n" + string(apiResponse.Body))
	fmt.Println("Bucket 0 Name: " + (allBuckets.Bucket[0].BucketName))
	fmt.Println("\n^ List Buckets Test Completed\n")
}

// Test createBucket
func TestToCreateBucket(t *testing.T) {
	B2CreateBucket("testbucket", false)
	fmt.Println("\n^ Create Bucket Test Completed\n")
}

// Test getUploadURL
func TestToReturnUploadURL(t *testing.T) {
	uploadResponse := B2GetUploadURL("b6ee61624837a6c6588b0715")
	fmt.Println(uploadResponse.URL)
	fmt.Println("\n^ Get Upload URL Test Completed\n")
}

// Test uploadFile
func TestToUploadFile(t *testing.T) {
	apiResponse := B2UploadFile("b6ee61624837a6c6588b0715",
		"/Users/dsjr2006/Dev/golang/src/github.com/dsjr2006/gopherb2/testfile.txt")
	fmt.Println(string(apiResponse.Body))
	fmt.Println("\n^ Upload File Test Completed\n")
}
/*
// Test B2StartLargeFile
func TestToStartLargeFile(t *testing.T) {
	apiResponse, fileResponse := B2StartLargeFile("b6ee61624837a6c6588b0715", "/Users/dsjr2006/Dev/golang/src/github.com/dsjr2006/gopherb2/rbsp_launch_1080p.mp4")
	fmt.Println(string(apiResponse.Body))
	fmt.Println("\nFile Name: " + fileResponse.FileName + "\nFile ID: " + fileResponse.FileID)
	fmt.Println("\n^ Start Large File Test Completed\n")
}

// Test B2GetUploadPartURL

func TestToGetUploadPartURL(t *testing.T) {
	apiResponse := B2GetUploadPartURL("4_zb6ee61624837a6c6588b0715_f202847f5820b0610_d20161214_m184330_c001_v0001036_t0053")
	fmt.Println("\nUpload Part URL: " + apiResponse.UploadURL)
	fmt.Println("\n^ Get Upload Part URL Test Completed (*Test FileID hardcoded)\n")
}
*/

// Test B2LargeFileUpload
func TestToUploadLargeFile(t *testing.T) {
	apiRespone, err := B2LargeFileUpload("b6ee61624837a6c6588b0715", "/Users/dsjr2006/Dev/golang/src/github.com/dsjr2006/gopherb2/rbsp_launch_1080p.mp4")
	if err != nil {
		fmt.Println("Fatal error %v", err)
	}
	fmt.Println("\nLarge Upload Response Body: *Nothing right now\n" + string(apiRespone.Body))
	fmt.Println("\n^ Upload Large File Test Completed\n")
}
