package gopherb2

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dsjr2006/blake2b-simd"
	"github.com/uber-go/zap"
)

func createTempFiles(undividedFile LargeFile) (LargeFile, error) {
	logger.Info("Creating Temp files from original large file",
		zap.String("File Path", undividedFile.OrigPath),
	)
	// Open undivided original file
	file, err := os.Open(undividedFile.OrigPath)
	defer file.Close()
	if err != nil {
		logger.Fatal("Cannot open undivided large file to create temp files.",
			zap.Error(err),
		)
	}
	// Get File Stats
	fileInfo, err := file.Stat()
	if err != nil {
		logger.Fatal("Cannot get stats of undivided large file.",
			zap.Error(err),
		)
	}
	// Check File Size Greater than 100MB Base-12 (102400 KB)
	// API doc recommends standard upload for < 200 MB, but min part size other than last is 100 MB
	if fileInfo.Size() < 104857600 {
		logger.Fatal("Large File is less than 100MB, use standard upload",
			zap.String("File Size", string(fileInfo.Size())),
		)
	}
	fileExtension := filepath.Ext(undividedFile.OrigPath)
	var fileSize int64 = fileInfo.Size()
	const fileChunk = 100 * (1 << 20) // 100 MB, change this to your requirement
	// calculate total number of parts the file will be chunked into
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	logger.Info("Splitting file into temp pieces",
		zap.Uint64("Number of Parts", totalPartsNum),
	)
	undividedFile.Pieces = int(totalPartsNum)
	if totalPartsNum > 10000 {
		logger.Fatal("File cannot be split into more than 10000 pieces")
	}
	// Process parts
	for i := uint64(0); i < totalPartsNum; i++ {
		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)
		file.Read(partBuffer)
		// Add trailing number to filename before extension "filename_1.ext" then Write to Disk
		tempFileName := os.TempDir() + strings.TrimSuffix(undividedFile.Name, fileExtension) + "_" + strconv.FormatUint(i, 10) + fileExtension
		_, err := os.Create(tempFileName)
		//tempFile, err := ioutil.TempFile("", tempFileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// write/save buffer to disk
		ioutil.WriteFile(tempFileName, partBuffer, os.ModeAppend)
		//tempFile.Write(partBuffer)
		// Get Temp file hash
		fileHash, err := fileSHA1(tempFileName)
		logger.Info("Temp File Piece Created",
			zap.Int("Piece #", int(i)),
			zap.String("Piece Filename", tempFileName),
			zap.String("Piece SHA1", fileHash),
		)

		uploadPartResponse := B2GetUploadPartURL(undividedFile.FileID)
		if uploadPartResponse.FileID != undividedFile.FileID {
			logger.Error("Upload Part File ID and Start File ID Do Not Match",
				zap.String("Part File ID", uploadPartResponse.FileID),
				zap.String("Start File ID", undividedFile.FileID),
			)
		}
		tempPiece := TempPiece{
			OrigFilePath:       undividedFile.OrigPath,
			OrigFileName:       undividedFile.Name,
			PieceNum:           int(i),
			SHA1:               fileHash,
			Size:               int64(partSize),
			Path:               tempFileName,
			URL:                uploadPartResponse.UploadURL,
			AuthorizationToken: uploadPartResponse.AuthorizationToken,
			FileID:             uploadPartResponse.FileID,
			UploadStatus:       "Not Started",
		}
		undividedFile.Temp = append(undividedFile.Temp, tempPiece)
	}
	return undividedFile, err
}

func removeTempFiles(largeFile LargeFile) {

	for i := 0; i < len(largeFile.Temp); i++ {
		if largeFile.Temp[i].UploadStatus != "Success" {
			logger.Error("Some temp files in large file were not uploaded",
				zap.String("Large file name", largeFile.Name),
				zap.String("Piece Path", largeFile.Temp[i].Path),
			)
		}
		if largeFile.Temp[i].UploadStatus == "Success" {
			os.Remove(largeFile.Temp[i].Path) // If Upload was successful remove file
			largeFile.Temp[i].UploadStatus = "Success - Deleted"
			logger.Info("Temporary File Deleted",
				zap.String("Large file name", largeFile.Name),
				zap.String("Piece Path", largeFile.Temp[i].Path),
			)
		}
	}
	return
}

func fileSHA1(filePath string) (string, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		logger.Warn("Could not retrieve file for SHA1 hashing",
			zap.Error(err),
		)
		return "fail", err
	}

	hash := sha1.New()
	// Copy file into hash interface
	if _, err := io.Copy(hash, file); err != nil {
		logger.Warn("File SHA1 Hash Failure",
			zap.Error(err),
		)
		return "fail", err
	}
	// Get 20 bytes hash
	hashAsBytes := hash.Sum(nil)[:20]

	return hex.EncodeToString(hashAsBytes), err
}

func fileBlake2b(filePath string) (string, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		logger.Warn("Could not retrieve file for Blake2b hashing",
			zap.Error(err),
		)
		return "fail", err
	}

	hash := blake2b.New512()
	// Copy file to hash interface
	if _, err := io.Copy(hash, file); err != nil {
		logger.Warn("File Blake2b Hash Failure",
			zap.Error(err),
		)
		return "fail", err
	}
	hashAsBytes := hash.Sum(nil)
	return hex.EncodeToString(hashAsBytes), err
}

func encodeFilename(filePath string) string {
	file, err := os.Open(filePath)
	defer file.Close()
	checkError(err)
	fileInfo, err := file.Stat()
	checkError(err)
	encodedFilename := string(fileInfo.Name())
	fmt.Println(encodedFilename)
	return encodedFilename
}
