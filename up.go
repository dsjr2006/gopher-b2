package gopherb2

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"os"
	"sync"
	"time"

	blake2b "github.com/dsjr2006/blake2b-simd"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// UploadConcurrency controls the number of simulataneous upload tasks of a multi-part upload
var UploadConcurrency = 5

type UpToB2File struct {
	Filepath      string
	Filename      string
	LastModMillis int64
	PieceSize     int64
	TotalSize     int64
	Blake2b       string
	SHA1          string
	Piece         []B2FilePiece // For B2 Large File - First Piece [0] will have Size/Hashes/Status
}
type B2FilePiece struct {
	PieceNum int
	SHA1     string
	Size     int64
	Status   string
}

func NewB2File(path string) (UpToB2File, error) {
	var b2F UpToB2File
	b2F.Filepath = path
	// Open undivided original file
	file, err := os.Open(b2F.Filepath)
	defer file.Close()
	if err != nil {
		return b2F, err
	}
	// Get File Stats
	fileInfo, err := file.Stat()
	if err != nil {
		return b2F, err
	}
	// Get File Modification Time as int64 value in milliseconds since midnight, January 1, 1970 UTC
	b2F.LastModMillis = fileInfo.ModTime().UnixNano() / 1000000
	b2F.TotalSize = fileInfo.Size()
	b2F.Filename = fileInfo.Name()

	const fileChunk int64 = 100 * (1 << 20) // 100 MB, change this to your requirement
	if fileChunk > b2F.TotalSize {
		b2F.PieceSize = b2F.TotalSize
	} else {
		b2F.PieceSize = fileChunk
	}

	// calculate total number of parts the file will be chunked into
	totalPartsNum := uint64(math.Ceil(float64(b2F.TotalSize) / float64(fileChunk)))
	fmt.Printf("\nTotal Parts Num: %v", totalPartsNum)
	if totalPartsNum > 10000 {
		//TODO increase chunk size if too many parts, will fail at too many pieces because of file size 1TB?
	}
	totalSize := b2F.TotalSize
	for i := 0; i < int(totalPartsNum); i++ {
		// Set piece size to calculated part size unless last piece
		var pieceSize int64
		if totalSize > b2F.PieceSize {
			pieceSize = b2F.PieceSize
		} else {
			pieceSize = totalSize
		}
		fmt.Printf("\nPiece size: %v", pieceSize)

		piece := B2FilePiece{
			Status:   "Unprocessed",
			Size:     pieceSize,
			PieceNum: i,
		}
		totalSize -= b2F.PieceSize
		fmt.Printf("\nUpdating Status of Piece# %v", i+1)
		b2F.Piece = append(b2F.Piece, piece)
	}
	err = b2F.Process()
	if err != nil {
		return b2F, err
	}

	return b2F, nil
}

// Process runs functions to get necessary file hashes, currently run at end of NewB2File
func (b2F *UpToB2File) Process() error {
	err := b2F.getPieceSHA1s()
	if err != nil {
		//TODO:Handle error
		return err
	}
	err = b2F.getSHA1()
	if err != nil {
		//TODO: Handle error
		return err
	}
	err = b2F.getBlakeb2()
	if err != nil {
		//TODO:Handle error
		return err
	}
	fmt.Printf("\nTotal Size: %v", b2F.getTotalSize())
	return nil
}

// Upload transmits file(s) to Backblaze B2
func (b2F *UpToB2File) Upload(bucketID string) error {
	// Standard Upload if one piece
	if len(b2F.Piece) == 1 {
		fmt.Println("Starting Standard upload")
		uploadURL := B2GetUploadURL(bucketID)
		file, err := os.Open(b2F.Filepath)
		if err != nil {
			//TODO: handle error
		}
		defer file.Close()
		// Create and Start Progress Bar
		pbar := pb.New(int(b2F.TotalSize)).SetUnits(pb.U_BYTES)
		pbar.SetRefreshRate(time.Second)
		pbar.ShowSpeed = true
		pbar.ShowTimeLeft = true
		pbar.Start()
		// Create and Send Request
		client := &http.Client{}
		req, err := http.NewRequest("POST", uploadURL.URL, pbar.NewProxyReader(file))
		req.ContentLength = b2F.TotalSize
		req.Header.Add("Authorization", uploadURL.AuthorizationToken)
		req.Header.Add("Content-Type", "b2/x-auto")
		req.Header.Add("X-Bz-Content-Sha1", b2F.SHA1)
		req.Header.Add("X-Bz-File-Name", b2F.Filename)
		req.Header.Add("X-Bz-Info-src_last_modified_millis", fmt.Sprintf("%d", b2F.LastModMillis))
		req.Header.Add("X-Bz-Info-Content-Blake2b", b2F.Blake2b)
		if err != nil {
			log.Fatalf("\nRequest failed. Error: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("\nResponse read fail. Error: %v", err)
		}
		pbar.Finish()
		// Read Response Body
		respBody, _ := ioutil.ReadAll(resp.Body)

		// Check API Response
		if resp.Status == "200 OK" {
			var uploaded UploadedFile
			err = json.Unmarshal(respBody, &uploaded)

			if uploaded.ContentSha1 != b2F.SHA1 {
				log.Fatal("API Response SHA1 Hash Mismatch.")
			}

			b2F.Piece[0].Status = "Upload Success"
			fmt.Printf("\nUpload Complete \nFilename: %v \nFileID: %v\n", uploaded.FileName, uploaded.FileID)

			return nil
		}
		if resp.Status != "200 OK" {
			requestDump, err := httputil.DumpRequest(req, true)
			if err != nil {
				log.Fatal("Could not dump HTTP request")
			}
			fmt.Printf("\nRequest: %v\n", string(requestDump))

			responseDump, err := httputil.DumpResponse(resp, true)
			if err != nil {
				log.Fatal("Could not dump HTTP response")
			}
			fmt.Printf("\nResponse: %v\n", string(responseDump))

			return errors.New("could not complete upload, please see log and retry")

		}
	}
	// Multi-part Upload if greather than one piece
	fmt.Println("Starting multi-part upload")
	file, err := os.Open(b2F.Filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	// Send start request to API and check response
	b2StartLgFile, err := b2F.startB2LargeFile(bucketID)
	if err != nil {
		log.Fatal("Start large file failed", err)
	}
	fmt.Printf("%v", b2StartLgFile) // TODO: Remove this
	// create progress bar pool
	pbpool, err := pb.StartPool()
	if err != nil {
		logger.Fatal("Could not start Progress Bar pool")
	}
	// create task channel
	filePieces := make(chan B2FilePiece)
	go func() {
		for i := 0; i < len(b2F.Piece); i++ {
			filePieces <- b2F.Piece[i]
		}
		close(filePieces)
	}()

	// waitgroup, and close results channel when done
	results := make(chan string)
	var wg sync.WaitGroup
	wg.Add(UploadConcurrency)
	go func() {
		wg.Wait()
		close(results)
	}()

	for i := 0; i < UploadConcurrency; i++ {
		go func(id int) {
			defer wg.Done()

			for p := range filePieces {
				// Progress Bar
				pbar := pb.New64(p.Size).SetUnits(pb.U_BYTES)
				pbar.SetRefreshRate(time.Second)
				pbar.Prefix(fmt.Sprintf("Part %v of %v", p.PieceNum+1, len(b2F.Piece)))
				pbar.ShowSpeed = true
				pbar.ShowTimeLeft = true
				pbpool.Add(pbar)

				// Get Upload Part URL & AuthorizationToken
				time.Sleep(time.Second)
				// Attempt Upload
				time.Sleep(10 * time.Second)
				fmt.Printf("Thread #%v Piece ID: %v Size: %v SHA1: %v\n", id, p.PieceNum, p.Size, p.SHA1)

				results <- "done"
			}
		}(i)
	}

	// loop over results until closed (see above)
	for r := range results {
		fmt.Printf("%v\n", r)
	}
	pbpool.Stop()

	return nil
}
func (b2F *UpToB2File) getTotalSize() int64 {
	var tSz int64
	// If Total Size not empty return total size
	if b2F.TotalSize != 0 {
		return b2F.TotalSize
	}
	for i := 0; i < len(b2F.Piece); i++ {
		tSz += b2F.Piece[i].Size
	}
	b2F.TotalSize = tSz
	return tSz
}
func (b2F *UpToB2File) getPieceSHA1s() error {
	file, err := os.Open(b2F.Filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	for i := 0; i < len(b2F.Piece); i++ {
		// Create buffer and fill buffer from file
		partBuffer := make([]byte, b2F.Piece[i].Size)
		_, err := file.Read(partBuffer)
		if err != nil {
			return err
		}

		hash := sha1.New()
		_, err = hash.Write(partBuffer)
		if err != nil {
			return err
		}
		// Get 20 bytes hash
		hashAsBytes := hash.Sum(nil)[:20]
		b2F.Piece[i].SHA1 = hex.EncodeToString(hashAsBytes)
	}

	return nil
}
func (b2F *UpToB2File) getSHA1() error {
	file, err := os.Open(b2F.Filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create buffer and fill buffer from file
	partBuffer := make([]byte, b2F.TotalSize)
	ptSz, err := file.Read(partBuffer)
	if err != nil {
		return err
	}
	if int64(ptSz) != b2F.TotalSize {
		//return errors.New("File Size read into buffer does not match file total size")
	}

	hash := sha1.New()
	_, err = hash.Write(partBuffer)
	if err != nil {
		return err
	}
	// Get 20 bytes hash
	hashAsBytes := hash.Sum(nil)[:20]
	b2F.SHA1 = hex.EncodeToString(hashAsBytes)

	return nil
}
func (b2F *UpToB2File) getBlakeb2() error {
	file, err := os.Open(b2F.Filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	partBuffer := make([]byte, b2F.TotalSize)
	file.Read(partBuffer)
	hash := blake2b.New512()
	_, err = hash.Write(partBuffer)
	if err != nil {
		return err
	}
	//io.WriteString(hash, string(partBuffer))
	// 32 byte hash
	hashAsBytes := hash.Sum(nil)[:32]
	b2F.Blake2b = hex.EncodeToString(hashAsBytes)

	return nil
}
func (b2F *UpToB2File) startB2LargeFile(bucketID string) (B2File, error) {
	// Authorize
	apiAuth := B2AuthorizeAccount()

	// Create client
	client := &http.Client{}
	// Request Body : JSON object
	jsonBody := []byte(`{"fileInfo": {"large_file_sha1": "` + b2F.SHA1 + `","src_last_modified_millis": "` + fmt.Sprintf("%d", b2F.LastModMillis) + `"},"bucketId": "` + bucketID + `","fileName": "` + b2F.Filename + `","contentType": "b2/x-auto"}`)
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
			log.Fatalf("File JSON parse failed. Error: %v", err)
		}
		return b2File, nil
	}
	return b2File, err
}
