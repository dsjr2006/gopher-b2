package gopherb2

import (
	"fmt"
	"testing"
)

// TestToReturnNewB2File does that
func TestToReturnNewB2File(t *testing.T) {
	b2F, err := NewB2File("/Users/dsjr2006/Dev/golang/src/github.com/dsjr2006/gopherb2/testfile.txt")
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	for i := 0; i < len(b2F.Piece); i++ {
		fmt.Printf("\nFile Piece %v- Size: %v - SHA1: %v", i, b2F.Piece[i].Size, b2F.Piece[i].SHA1)
	}
	fmt.Printf("\nFile Blake2b: %v", b2F.Blake2b)
	fmt.Println("\n^ New B2 File Test Completed\n")
	return
}

// TestToReturnNewB2File does that
func TestToUploadNewStandardB2File(t *testing.T) {
	b2F, err := NewB2File("/Users/dsjr2006/Dev/golang/src/github.com/dsjr2006/gopherb2/testfile.txt")
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	for i := 0; i < len(b2F.Piece); i++ {
		fmt.Printf("\nFile Piece %v- Size: %v - SHA1: %v", i, b2F.Piece[i].Size, b2F.Piece[i].SHA1)
	}
	err = b2F.Upload("b6ee61624837a6c6588b0715")
	if err != nil {
		fmt.Printf("Could not upload file. Error: %v", err)
	}
	fmt.Printf("\nFile Blake2b: %v", b2F.Blake2b)
	fmt.Println("\n^ Standard B2 File Upload Test Completed\n")
	return
}

// TestToReturnNewLargeB2File
func TestToReturnNewLargeB2File(t *testing.T) {
	//b2F, err := NewB2File("/Users/dsjr2006/Downloads/megan@allaboutent.com.zip") // ~ 3GB
	b2F, err := NewB2File("/Users/dsjr2006/Downloads/LibreOffice_5.3.0_MacOS_x86-64.dmg") // ~ 250MB
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	for i := 0; i < len(b2F.Piece); i++ {
		fmt.Printf("\nFile Piece %v- Size: %v - SHA1: %v", i, b2F.Piece[i].Size, b2F.Piece[i].SHA1)
	}
	fmt.Printf("\nFile Blake2b: %v", b2F.Blake2b)
	fmt.Println("\n^ New Large B2 File Test Completed\n")

	fmt.Println("Upload Test..")
	err = b2F.Upload("b6ee61624837a6c6588b0715")
	if err != nil {
		fmt.Printf("Could not upload file. Error: %v", err)
	}
	return
}

// Test authorizeAccount
func TestToReturnAuthorization(t *testing.T) {
	apiResponse := B2AuthorizeAccount()
	fmt.Println("\nAccount ID: " + apiResponse.AccountID)
	fmt.Println("\n^ Authorization Test Completed\n")
}

// Test listBuckets
func TestToReturnBucketList(t *testing.T) {
	B2ListBuckets()
	//fmt.Println("\nBuckets Response Body:\n" + string(apiResponse.Body))
	fmt.Println("\n\n^ List Buckets Test Completed\n")
}

// Test B2ListFilenames
func TestToReturnFilenames(t *testing.T) {
	B2ListFilenames("b6ee61624837a6c6588b0715", "")

	fmt.Println("\n\n^ List Filenames complete")
}

/* Test createBucket
func TestToCreateBucket(t *testing.T) {
	B2CreateBucket("testbucket", false)
	fmt.Println("\n^ Create Bucket Test Completed\n")
}
*/

// Test getUploadURL
func TestToReturnUploadURL(t *testing.T) {
	uploadResponse := B2GetUploadURL("b6ee61624837a6c6588b0715")
	fmt.Printf("\nUpload URL Received: %v", uploadResponse.URL)
	fmt.Println("\n^ Get Upload URL Test Completed\n")
}

// Test uploadFile
func TestToUploadFile(t *testing.T) {
	UploadFile("b6ee61624837a6c6588b0715",
		"/Users/dsjr2006/Dev/golang/src/github.com/dsjr2006/gopherb2/testfile.txt")
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


// Test B2LargeFileUpload
func TestToUploadFile(t *testing.T) {
	UploadFile("b6ee61624837a6c6588b0715", "/Users/dsjr2006/Dev/golang/src/github.com/dsjr2006/gopherb2/rbsp_launch_1080p.mp4")

	fmt.Println("\n^ Upload Large File Test Completed\n")
}
*/
