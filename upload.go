package gopherb2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/uber-go/zap"
)

type LargeFile struct {
	Name                   string
	SHA1                   string
	Pieces                 int
	OrigPath               string
	Temp                   []TempPiece
	LastModificationMillis int64
	FileID                 string
	Size                   int64
}
type TempPiece struct {
	OrigFilePath       string
	OrigFileName       string
	PieceNum           int
	SHA1               string
	Size               int64
	Path               string
	URL                string
	AuthorizationToken string
	FileID             string
	UploadStatus       string
}
type UploadURL struct {
	AuthorizationToken string `json:"authorizationToken"`
	BucketId           string `json:"bucketId"`
	URL                string `json:"uploadUrl"`
}

// Upload single file to B2
/*
TODO: Check SHA1 match after upload
TODO: Verbose errors
TODO: Info Log
*/
func UploadFile(bucketId string, filePath string) {
	// Determine Upload Method
	file, err := os.Stat(filePath)

	// defer file.Close()
	checkError(err)
	/*
	fileInfo, err := file.Stat()
	checkError(err)
	*/

	if file.Size() < 104857600 {
		B2UploadFile(bucketId, filePath)
	} else {
		B2LargeFileUpload(bucketId, filePath)
	}

	return
}
func B2UploadFile(bucketId string, filePath string) {
	// Authorize and Get Upload URL
	uploadURL := B2GetUploadURL(bucketId)

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
		logger.Fatal("Error creating upload buffer",
			zap.Error(err),
		)
	}
	checkError(err)
	body := buffer
	// Get File Modification Time as int64 value in milliseconds since midnight, January 1, 1970 UTC
	fileModTimeMillis := fileInfo.ModTime().UnixNano() / 1000000

	// Create request
	req, err := http.NewRequest("POST", uploadURL.URL, body)
	// TODO: Handle Request Error

	// Get File Hash
	fileHash, err := fileSHA1(filePath)
	fileBlake2b, err := fileBlake2b(filePath)
	// TODO: Handle File Hash Return Error?

	// Headers
	req.Header.Add("Authorization", uploadURL.AuthorizationToken)
	req.Header.Add("Content-Type", "b2/x-auto")
	req.Header.Add("Content-Length", string(fileInfo.Size()))
	req.Header.Add("X-Bz-Content-Sha1", fileHash)
	req.Header.Add("X-Bz-File-Name", fileInfo.Name()) //Need to encode names properly! according to B2 docs
	req.Header.Add("X-Bz-Info-src_last_modified_millis", fmt.Sprintf("%d", fileModTimeMillis))
	req.Header.Add("X-Bz-Info-Content-Blake2b", fileBlake2b)
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
	if apiResponse.Status == "200 OK" {
		logger.Info("Upload File Sucessful",
			zap.String("Filename:", filePath),
		)
	} else {
		logger.Panic("Could not upload file",
			zap.String("API Resp Body:", string(apiResponse.Body)),
		)
	}
	return
}

func B2LargeFileUpload(bucketId string, filePath string) {
	// Open File and Get File Stats
	file, err := os.Open(filePath)
	defer file.Close()
	// TODO: Check file open error
	fileInfo, err := file.Stat()
	// TODO: Check file stat error

	// Send start request to API and check response
	startResp, b2File := B2StartLargeFile(bucketId, filePath)
	if startResp.Status != "200 OK" {
		logger.Warn("Invalid response to start large file request",
			zap.String("Response", string(startResp.Body)),
		)
	}
	var largeFile LargeFile
	largeFile.Name = fileInfo.Name()
	largeFile.OrigPath = filePath
	largeFile.LastModificationMillis = b2File.FileInfo.LastModificationMillis
	largeFile.FileID = b2File.FileID
	largeFile.Size = fileInfo.Size()
	sha1, err := fileSHA1(filePath)

	// TODO: Check SHA1 error
	largeFile.SHA1 = sha1

	largeFile, err = createTempFiles(largeFile)
	if err != nil {
		logger.Fatal("Invalid response from create temp files operation.",
			zap.Error(err),
		)
	}
	// Do simultaneous multipart upload
	logger.Info("Beginning Multipart Upload",
		zap.String("B2 File ID", largeFile.FileID),
		zap.Int64("Size", largeFile.Size),
		zap.Int("Pieces", largeFile.Pieces),
	)
	uploadParts(largeFile)
	if err != nil {
		logger.Fatal("Upload Parts of Large File failed",
			zap.Error(err),
		)
	}
	err = B2FinishLargeFile(largeFile)

	if err != nil {
		logger.Warn("Could not complete large file",
			zap.Error(err),
		)
	}
	if err == nil {
		removeTempFiles(largeFile)
	}

	return
}

// Begin Large File Upload
func B2StartLargeFile(bucketId string, filePath string) (Response, B2File) {
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
	largeFileSHA1, err := fileSHA1(filePath)
	if largeFileSHA1 == "fail" {
		logger.Fatal("Cannot parse API Auth Response JSON.")
	}

	// Create client
	client := &http.Client{}
	// Request Body : JSON object
	jsonBody := []byte(`{"fileInfo": {"large_file_sha1": "` + largeFileSHA1 + `","src_last_modified_millis": "` + fmt.Sprintf("%d", fileModTimeMillis) + `"},"bucketId": "` + bucketId + `","fileName": "` + fileInfo.Name() + `","contentType": "b2/x-auto"}`)
	body := bytes.NewBuffer(jsonBody)

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

	// Parse API Response File Info to B2File if request is successful
	var b2File B2File
	if apiResponse.Status == "200 OK" {
		err = json.Unmarshal(apiResponse.Body, &b2File)
		if err != nil {
			logger.Fatal("File JSON Parse Failed",
				zap.Error(err),
			)
		}
		return apiResponse, b2File
	}

	return apiResponse, b2File
}

func B2UploadPart(largeFile LargeFile, pieceNum int, wg *sync.WaitGroup) {
	defer wg.Done()
	logger.Info("Starting Upload of Part",
		zap.Int("B2 Part #", pieceNum+1),
		zap.String("Piece Path", largeFile.Temp[pieceNum].Path),
		zap.String("Upload URL", largeFile.Temp[pieceNum].URL),
	)
	file, err := os.Open(largeFile.Temp[pieceNum].Path)
	defer file.Close()
	// Upload Part (POST https://pod-000-1004-12.backblaze.com/b2api/v1/b2_upload_part/....)
	// Create client
	client := &http.Client{}

	// TODO: Read Request Body from Disk for lower memory usage?
	// Request Body
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		logger.Fatal("Could not create part upload buffer",
			zap.Error(err),
		)
	}
	checkError(err)
	body := buffer

	// Create request
	req, err := http.NewRequest("POST", largeFile.Temp[pieceNum].URL, body)

	// Headers
	req.Header.Add("Content-Length", string(largeFile.Temp[pieceNum].Size))
	req.Header.Add("X-Bz-Part-Number", fmt.Sprintf("%d", (pieceNum+1))) // Temp files begin at 0, increase by 1 to match B2 response
	req.Header.Add("Authorization", largeFile.Temp[pieceNum].AuthorizationToken)
	req.Header.Add("X-Bz-Content-Sha1", largeFile.Temp[pieceNum].SHA1)

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := ioutil.ReadAll(resp.Body)

	var apiResponse Response
	apiResponse.Header = resp.Header
	apiResponse.Status = resp.Status
	apiResponse.Body = respBody
	if apiResponse.Status == "200 OK" {
		logger.Info("Part Uploaded Successfully",
			zap.String("Filename", largeFile.Name),
			zap.Int("B2 Piece Num", int(pieceNum+1)),
			zap.String("Piece Path", largeFile.Temp[pieceNum].Path),
			zap.String("Respone Body", string(apiResponse.Body)),
		)
		largeFile.Temp[pieceNum].UploadStatus = "Success"
	}
	if apiResponse.Status != "200 OK" {
		logger.Warn("Part Upload Failed",
			zap.String("Filename", largeFile.Name),
			zap.Int("B2 Piece Num", int(pieceNum+1)),
			zap.String("Piece Path", largeFile.Temp[pieceNum].Path),
			zap.String("Response Body", string(apiResponse.Body)),
		)
		largeFile.Temp[pieceNum].UploadStatus = "Failed"
	}

	return
}
func uploadParts(largeFile LargeFile) {
	var wg sync.WaitGroup
	logger.Info("Uploading File Part",
		zap.String("File", largeFile.Name),
	)
	for i := 0; i < len(largeFile.Temp); i++ {
		wg.Add(1)
		go B2UploadPart(largeFile, i, &wg)
	}
	wg.Wait()

	return
}
