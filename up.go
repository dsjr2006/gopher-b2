package gopherb2

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math"
	"os"

	blake2b "github.com/dsjr2006/blake2b-simd"
)

type UpToB2File struct {
	Filepath      string
	LastModMillis int64
	PieceSize     int64
	Piece         []B2FilePiece // For B2 Large File - First Piece [0] will have Size/Hashes/Status
}
type B2FilePiece struct {
	SHA1    string
	Blake2b string
	Size    int64
	Status  string
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
	var fileSize = fileInfo.Size()
	fmt.Printf("\nFile Size: %v", fileSize)

	const fileChunk int64 = 100 * (1 << 20) // 100 MB, change this to your requirement
	if fileChunk > fileSize {
		b2F.PieceSize = fileSize
	} else {
		b2F.PieceSize = fileChunk
	}

	// calculate total number of parts the file will be chunked into
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	fmt.Printf("\nTotal Parts Num: %v", totalPartsNum)
	if totalPartsNum > 10000 {
		//TODO increase chunk size if too many parts, will fail at too many pieces because of file size 1TB?
	}
	for i := 0; i < int(totalPartsNum); i++ {
		piece := B2FilePiece{
			Status: "Unprocessed",
		}
		fmt.Printf("\nUpdating Status of Piece# %v", i+1)
		b2F.Piece = append(b2F.Piece, piece)
	}
	err = b2F.Process()
	if err != nil {
		return b2F, err
	}

	return b2F, nil
}

func (b2F *UpToB2File) Process() error {
	err := b2F.getAllsha1()
	if err != nil {
		//TODO:Handle error
		return err
	}
	err = b2F.getAllblakeb2()
	if err != nil {
		//TODO:Handle error
		return err
	}
	fmt.Printf("\nTotal Size: %v", b2F.getTotalSize())
	return nil
}
func (b2F *UpToB2File) getTotalSize() int64 {
	var tSz int64
	for i := 0; i < len(b2F.Piece); i++ {
		tSz += b2F.Piece[i].Size
	}
	return tSz
}
func (b2F *UpToB2File) getAllsha1() error {
	file, err := os.Open(b2F.Filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	for i := 0; i < len(b2F.Piece); i++ {
		fmt.Printf("\nPiece size: %v", b2F.PieceSize)
		// Create buffer and fill buffer from file
		partBuffer := make([]byte, b2F.PieceSize)
		ptSz, err := file.Read(partBuffer)
		if err != nil {
			return err
		}
		b2F.Piece[i].Size = int64(ptSz)

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

func (b2F *UpToB2File) getAllblakeb2() error {
	file, err := os.Open(b2F.Filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	for i := 0; i < len(b2F.Piece); i++ {
		partBuffer := make([]byte, b2F.PieceSize)
		file.Read(partBuffer)
		hash := blake2b.New512()
		_, err := hash.Write(partBuffer)
		if err != nil {
			return err
		}
		//io.WriteString(hash, string(partBuffer))
		// 32 byte hash
		hashAsBytes := hash.Sum(nil)[:32]
		b2F.Piece[i].Blake2b = hex.EncodeToString(hashAsBytes)
	}

	return nil
}
