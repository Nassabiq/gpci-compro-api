package domain

import "errors"

var (
	ErrEmptyFile       = errors.New("file is empty")
	ErrUnsupportedType = errors.New("file type is not allowed")
	ErrNoFilesProvided = errors.New("no files provided")
)

type UploadResult struct {
	Directory        string `json:"directory"`
	Filename         string `json:"filename"`
	OriginalFilename string `json:"original_filename"`
	MimeType         string `json:"mime_type"`
	Size             int64  `json:"size"`
	Category         string `json:"category"`
}
