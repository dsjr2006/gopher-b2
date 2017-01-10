package gopherb2
// TODO: Read uploads from disk to reduce memory usage? optional?
// TODO: Store Auth Info to reduce API requests, use old and re-auth if needed?
// TODO: Check if existing file?
// TODO: Check if unfinished large file?
// TODO: Increase chunk size to reduce number of uploads?
// TODO: Retry failed uploads/parts
// TODO: Add ability to set logging level
// TODO: Log to file
// TODO: Upload progress
// TODO: Upload timeout?
// TODO: Automatically select standard or large file upload
// TODO: Organize package
// TODO: Check for success on all files or resend, timeout? num of tries?
// TODO: Limit number of simultaneous uploads
// TODO: Encrypt files before upload
import (
	"net/http"

	"github.com/uber-go/zap"
)

// Setup Logging
var logger = zap.New(
	zap.NewJSONEncoder(),
)

func init()	{

}

type Configuration struct {
	ACCOUNT_ID     string
	APPLICATION_ID string
	API_URL        string
}
type Response struct {
	Header http.Header
	Status string
	Body   []byte
}
type APIAuthorization struct {
	AccountID          string `json:"accountId"`
	ApiURL             string `json:"apiUrl"`
	AuthorizationToken string `json:"authorizationToken"`
	DownloadURL        string `json:"downloadURL"`
	MinimumPartSize    int    `json:"minimumPartSize"`
}
type Buckets struct {
	Bucket []struct {
		AccountID      string   `json:"accountId"`
		BucketID       string   `json:"bucketId"`
		BucketName     string   `json:"bucketName"`
		BucketType     string   `json:"bucketType"`
		LifecycleRules []string `json:"lifecycleRules"`
		Revision       int      `json:"revision"`
	} `json:"buckets"`
}

type UploadPartResponse struct {
	AuthorizationToken string `json:"authorizationToken"`
	FileID             string `json:"fileId"`
	UploadURL          string `json:"uploadUrl"`
}
type B2File struct {
	AccountID   string `json:"accountId"`
	BucketID    string `json:"bucketId"`
	ContentType string `json:"contentType"`
	FileID      string `json:"fileId"`
	FileInfo    struct {
		LargeFileSHA1          string `json:"large_file_sha1"`
		LastModificationMillis int64  `json:"src_last_modified_millis,string"`
	} `json:"fileInfo"`
	FileName        string `json:"fileName"`
	UploadTimestamp int64  `json:"uploadTimestamp"`
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
	UploadStatus			 string
}
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

// Simple Error check for fatal errors
func checkError(e error) {
	if e != nil {
		logger.Fatal("checkError failed",
			zap.Error(e),
		)
		panic(e)
	}
}
