package gopherb2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/uber-go/zap"
	pb "gopkg.in/cheggaaa/pb.v1"
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
type UploadedFile struct {
	AccountID     string `json:"accountId"`
	Action        string `json:"action"`
	BucketID      string `json:"bucketId"`
	ContentLength int    `json:"contentLength"`
	ContentSha1   string `json:"contentSha1"`
	ContentType   string `json:"contentType"`
	FileID        string `json:"fileId"`
	FileInfo      struct {
		ContentBlake2B        string `json:"content-blake2b"`
		SrcLastModifiedMillis string `json:"src_last_modified_millis"`
	} `json:"fileInfo"`
	FileName        string `json:"fileName"`
	UploadTimestamp int64  `json:"uploadTimestamp"`
}

// UploadFile transmits file at given path to B2 Storage
func UploadFile(bucketID string, filePath string) error {
	// Determine Upload Method
	file, err := os.Stat(filePath)

	// defer file.Close()
	if err != nil {
		logger.Fatal("Unable to get file stats.",
			zap.Error(err),
		)
		log.Fatalf("Unable to get file stats. Error: %v", err)
	}

	if file.Size() < 120586240 {
		log.Debug("Sending file to Standard upload.")
		b2UploadStdFile(bucketID, filePath)
	} else {
		log.Debug("Sending file to Large upload")
		LargeFileUpload(bucketID, filePath)
	}

	return err
}
func b2UploadStdFile(bucketID string, filePath string) {
	b2F, err := NewB2File(filePath)
	if err != nil {
		log.Fatal("New B2 File Failure. ", err)
		return
	}
	err = b2F.Upload(bucketID)
	if err != nil {
		log.Fatal("Could not upload file", err)
	}
	return

}

func LargeFileUpload(bucketID string, filePath string) {
	// Open File and Get File Stats
	file, err := os.Open(filePath)
	defer file.Close()
	// TODO: Check file open error
	fileInfo, err := file.Stat()
	// TODO: Check file stat error

	// Send start request to API and check response
	startResp, b2File := B2StartLargeFile(bucketID, filePath)
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
func B2StartLargeFile(bucketID string, filePath string) (Response, B2File) {
	// Authorize
	apiAuth := AuthorizeAcct()

	// Open File and Get File Stats
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		log.Fatalf("Unable to open file. Error: %v", err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("Unable to get file stats. Error: %v", err)
	}
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
	jsonBody := []byte(`{"fileInfo": {"large_file_sha1": "` + largeFileSHA1 + `","src_last_modified_millis": "` + fmt.Sprintf("%d", fileModTimeMillis) + `"},"bucketId": "` + bucketID + `","fileName": "` + fileInfo.Name() + `","contentType": "b2/x-auto"}`)
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
func uploadParts(largeFile LargeFile) {
	var wg sync.WaitGroup
	logger.Info("Uploading File Part",
		zap.String("File", largeFile.Name),
	)
	pbpool, err := pb.StartPool()
	if err != nil {
		logger.Fatal("Could not start Progress Bar pool")
	}
	for i := 0; i < len(largeFile.Temp); i++ {
		wg.Add(1)
		go UploadPart(largeFile, i, &wg, pbpool)
	}
	wg.Wait()
	pbpool.Stop()

	return
}
func UploadPart(largeFile LargeFile, pieceNum int, wg *sync.WaitGroup, pbpool *pb.Pool) {
	defer wg.Done()
	logger.Info("Starting Upload of Part",
		zap.Int("B2 Part #", pieceNum+1),
		zap.String("Piece Path", largeFile.Temp[pieceNum].Path),
		zap.String("Upload URL", largeFile.Temp[pieceNum].URL),
	)
	file, err := os.Open(largeFile.Temp[pieceNum].Path)
	if err != nil {
		log.Fatalf("Unable to open file. Error: %v", err)
	}
	defer file.Close()
	// Upload Part (POST https://pod-000-1004-12.backblaze.com/b2api/v1/b2_upload_part/....)
	// Create client
	client := &http.Client{}

	// Progress Bar
	pbar := pb.New64(largeFile.Temp[pieceNum].Size).SetUnits(pb.U_BYTES)
	pbar.SetRefreshRate(time.Second)
	pbar.Prefix(fmt.Sprintf("Part %v of %v", pieceNum+1, len(largeFile.Temp)))
	pbar.ShowSpeed = true
	pbar.ShowTimeLeft = true

	pbpool.Add(pbar)

	// Create request
	req, err := http.NewRequest("POST", largeFile.Temp[pieceNum].URL, pbar.NewProxyReader(file))

	// Headers
	req.ContentLength = largeFile.Temp[pieceNum].Size
	req.Header.Add("X-Bz-Part-Number", fmt.Sprintf("%d", (pieceNum+1))) // Temp files begin at 0, increase by 1 to match B2 response
	req.Header.Add("Authorization", largeFile.Temp[pieceNum].AuthorizationToken)
	req.Header.Add("X-Bz-Content-Sha1", largeFile.Temp[pieceNum].SHA1)

	pbar.Start()

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}
	pbar.Finish()

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
			zap.String("Response Body", string(apiResponse.Body)),
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
