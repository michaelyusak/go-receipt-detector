package localstorage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	mimeToExt = map[string]string{
		"image/jpeg":      ".jpg",
		"image/png":       ".png",
		"image/gif":       ".gif",
		"image/webp":      ".webp",
		"image/bmp":       ".bmp",
		"image/tiff":      ".tiff",
		"image/svg+xml":   ".svg",
		"application/pdf": ".pdf",
	}
)

type receiptImages struct {
	localDirectory string
	serverBaseUrl  string
}

func NewReceiptImages(localDirectory, serverBaseUrl string) *receiptImages {
	return &receiptImages{
		localDirectory: localDirectory,
		serverBaseUrl:  serverBaseUrl,
	}
}

func (r *receiptImages) generateFilename(contentType string) (string, error) {
	ext, ok := mimeToExt[contentType]
	if !ok {
		return "", fmt.Errorf("[repository][localstorage][receiptImages][checkLocalDirectory] Unallowed content type")
	}

	year, month, day := time.Now().Date()
	dir := fmt.Sprintf("%s/%v-%v-%v", r.localDirectory, year, month, day)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("[repository][localstorage][receiptImages][checkLocalDirectory][os.MkdirAll] Failed to create directory: %w [dir: %s]", err, dir)
	}

	fileName := fmt.Sprintf("%s/%s%s", dir, uuid.New().String(), ext)

	return fileName, nil
}

func (r *receiptImages) StoreOne(ctx context.Context, contentType string, fileHeader *multipart.FileHeader) (string, error) {
	fileName, err := r.generateFilename(contentType)
	if err != nil {
		return "", fmt.Errorf("[repository][localstorage][StoreOne][r.generateFilename] Failed to generate file name : %w", err)
	}

	source, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("[repository][localstorage][StoreOne][fileHeader.Open] Failed to open source file: %w", err)
	}
	defer source.Close()

	out, err := os.Create(fileName)
	if err != nil {
		return "", fmt.Errorf("[repository][localstorage][StoreOne][os.Create] Failed to create out file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, source)
	if err != nil {
		return "", fmt.Errorf("[repository][localstorage][StoreOne][io.Copy] Failed to copy content to file: %w", err)
	}

	return fileName, nil
}

func (r *receiptImages) GetImageUrl(ctx context.Context, filePath string) (string, error) {
	url := strings.Replace(filePath, r.localDirectory, r.serverBaseUrl, 1)

	return url, nil
}
