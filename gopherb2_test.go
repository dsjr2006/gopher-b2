package gopherb2

import (
  "testing"
  "fmt"
)

// Test authorizeAccount
func TestToReturnAuthorization (t *testing.T) {
  apiResponse := authorizeAccount()
  fmt.Println("\nAccount ID: " + apiResponse.AccountID)
  fmt.Println("\n^ Authorization Test Completed\n")
}
// Test listBuckets
func TestToReturnBucketList (t *testing.T) {
  allBuckets := listBuckets()
  //fmt.Println("\nBuckets Response Body:\n" + string(apiResponse.Body))
  fmt.Println("Bucket 0 Name: " + (allBuckets.Bucket[0].BucketName))
  fmt.Println("\n^ List Buckets Test Completed\n")
}
// Test createBucket
func TestToCreateBucket (t *testing.T) {
  apiResponse := createBucket("testbucket",false)
  fmt.Println(string(apiResponse.Body))
  fmt.Println("\n^ Create Bucket Test Completed\n")
}
// Test getUploadURL
func TestToReturnUploadURL (t *testing.T) {
  uploadResponse := getUploadURL("b6ee61624837a6c6588b0715")
  fmt.Println(uploadResponse.URL)
  fmt.Println("\n^ Get Upload URL Test Completed\n")
}
