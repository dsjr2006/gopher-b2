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
  bucketList := listBuckets()
  fmt.Println("\nBuckets JSON:\n" + string(bucketList))
}
