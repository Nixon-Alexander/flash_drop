package helper

import (
	"errors"
	"flash_drop/types"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	MaxUploadSize = 10 << 20 // 10 MB
)

func UploadFile(c *gin.Context, files []*multipart.FileHeader) ([]types.File, error) {
	fmt.Println(len(files))
	if len(files) == 0 {
		return nil, errors.New("File required")
	}

	var results []types.File

	for _, file := range files {
		// Generate unique filename
		ext := filepath.Ext(file.Filename)
		newFilename := uuid.New().String() + ext
		dst := filepath.Join("./uploads", newFilename)
		fmt.Println(dst)

		// Save file
		if err := c.SaveUploadedFile(file, dst); err != nil {
			return nil, err
		}

		results = append(results, types.File{
			NewFileName:      newFilename,
			OriginalFileName: file.Filename,
			Size:             file.Size,
		})
	}

	return results, nil
}

func EditFile(c *gin.Context, oldFiles string, newFiles []*multipart.FileHeader) ([]types.File, error) {
	if len(newFiles) == 0 {
		return nil, errors.New("File required")
	}

	var results []types.File

	for _, file := range newFiles {
		// Delete old files
		oldDst := fmt.Sprint("./uploads/" + oldFiles)
		os.Remove(oldDst)

		// Generate unique filename
		ext := filepath.Ext(file.Filename)
		newFilename := uuid.New().String() + ext
		dst := filepath.Join("./uploads/", newFilename)

		// Save file
		if err := c.SaveUploadedFile(file, dst); err != nil {
			return nil, err
		}

		results = append(results, types.File{
			NewFileName:      newFilename,
			OriginalFileName: file.Filename,
			Size:             file.Size,
		})
	}

	return results, nil
}
