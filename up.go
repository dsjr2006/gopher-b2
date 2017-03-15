package gopherb2

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"time"

	pb "gopkg.in/cheggaaa/pb.v1"

	"log"

	blake2b "github.com/dsjr2006/blake2b-simd"
)

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
	SHA1   string
	Size   int64
	Status string
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
			Status: "Unprocessed",
			Size:   pieceSize,
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

// Process runs functions to get necessary file hashes
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
func (b2F *UpToB2File) Upload(bucketId string) error {
	// Standard Upload if one piece
	if len(b2F.Piece) == 1 {
		fmt.Println("Starting Standard upload")
		uploadURL := B2GetUploadURL(bucketId)
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

			fmt.Printf("\nUpload Complete \nFilename: %v \nFileID: %v\n", uploaded.FileName, uploaded.FileID)
		}
		return nil
	}
	// Multi-part Upload if greather than one piece
	fmt.Println("Starting multi-part upload")

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
		return errors.New("File Size read into buffer does not match file total size")
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
