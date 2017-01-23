package gopherb2

import (
    "fmt"
    "encoding/json"


    "gopkg.in/resty.v0"
	"github.com/uber-go/zap"
)

type allFiles struct {
	File []file `json:"files"`
	NextFileName string `json:"nextFileName"`
}
type file struct {
    Action string `json:"action"`
    ContentLength int `json:"contentLength"`
    ContentSha1 string `json:"contentSha1"`
    ContentType string `json:"contentType"`
    FileID string `json:"fileId"`
    FileInfo struct {
        ContentBlake2B string `json:"content-blake2b"`
        SrcLastModifiedMillis string `json:"src_last_modified_millis"`
    } `json:"fileInfo"`
    FileName string `json:"fileName"`
    Size int `json:"size"`
    UploadTimestamp int64 `json:"uploadTimestamp"`
}

// B2ListFilenames lists all files
func B2ListFilenames(bucketId string, startFile string) {
    // Authorize and Get API Token
	authorizationResponse := B2AuthorizeAccount()

    // Create & Fetch Request
    resp, err := resty.R().
        SetHeader("Authorization", authorizationResponse.AuthorizationToken).
        SetBody(`{"bucketId":"`+ bucketId +`","startFileName":"`+ startFile +`"}`).
        Post(authorizationResponse.ApiURL+"/b2api/v1/b2_list_file_names")
    if err != nil {
        logger.Fatal("API Communication Error: Could not get filename list",
            zap.Error(err),
        )
    }

    // Read Response
    var allFiles allFiles
    err = json.Unmarshal(resp.Body(), &allFiles)
    if err != nil {
        logger.Fatal("Error parsing JSON response for request all filenames",
            zap.Error(err),
        )
    }

    // Display files
    for i := 0; i < len(allFiles.File); i++ {
        fmt.Printf("\n\nFileID: %v\nFilename: %v\nSHA1: %v\nBlake2b: %v\nSize: %v", 
        allFiles.File[i].FileID, allFiles.File[i].FileName, allFiles.File[i].ContentSha1, 
        allFiles.File[i].FileInfo.ContentBlake2B, allFiles.File[i].Size)
    }

}