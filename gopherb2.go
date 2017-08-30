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
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/uber-go/zap"
)

// Log
// Log Level
var (
	LogDebug bool
	logger   = logLevel()
	LogDest  = ""
	Logger   = log.New()
)

type Configuration struct {
	AcctID string
	AppID  string
	APIURL string
}
type Response struct {
	Header http.Header
	Status string
	Body   []byte
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

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.WarnLevel)
}

func logLevel() zap.Logger {
	if LogDebug == true {
		return zap.New(
			zap.NewJSONEncoder(),
			zap.DebugLevel,
		)
	}

	return zap.New(
		zap.NewJSONEncoder(),
		zap.InfoLevel,
	)
}
func SetLogLevel(lvl string) error {
	switch lvl {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}
	return nil
}
