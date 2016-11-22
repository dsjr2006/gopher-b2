package gopher-b2

import (
  "testing"
  "github.com/dsjr2006/gopher-b2"
)

// Test authorizeAccount
func TestToReturnAuthorization (t *testing.T) {
  responseHeader, responseBody := authorizeAccount()
}
