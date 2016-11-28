package gopherb2

import (
  "testing"
  "fmt"
)

// Test authorizeAccount
func TestToReturnAuthorization (t *testing.T) {
  apiResponse := authorizeAccount()
  fmt.Println("\nAccount ID: " + apiResponse.AccountID)
}
// Test listBuckets
func TestToReturnBucketList (t *testing.T) {
  apiResponse := listBuckets()
  fmt.Println("\nBuckets JSON:\n" + string(apiResponse.Body))
}
// Test createBucket
func TestToCreateBucket (t *testing.T) {
  apiResponse := createBucket("testbucket",false)
  fmt.Println(string(apiResponse.Body))
}
