package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Nassabiq/gpci-compro-api/internal/modules/uploads/domain"
	"github.com/rs/xid"
)

type Service struct {
	repo        StorageRepository
	Bucket      string
	basePath    string
	modulePaths map[string]string
}

type StorageRepository interface {
	PutObject(ctx context.Context, bucket, objectName string, reader io.ReadSeeker, size int64, contentType string) error
}

func New(repo StorageRepository, bucket, basePath string, modulePaths map[string]string) *Service {
	cleanBase := strings.Trim(basePath, "/")
	normalized := make(map[string]string, len(modulePaths))
	for key, value := range modulePaths {
		k := strings.ToLower(strings.TrimSpace(key))
		if k == "" {
			continue
		}
		normalized[k] = strings.Trim(value, "/")
	}
	return &Service{
		repo:        repo,
		Bucket:      bucket,
		basePath:    cleanBase,
		modulePaths: normalized,
	}
}

func (s *Service) UploadFiles(ctx context.Context, module string, fileHeaders []*multipart.FileHeader) ([]domain.UploadResult, error) {
	if len(fileHeaders) == 0 {
		return nil, domain.ErrNoFilesProvided
	}

	module = strings.ToLower(strings.TrimSpace(module))

	results := make([]domain.UploadResult, 0, len(fileHeaders))
	for _, fh := range fileHeaders {
		result, err := s.uploadSingle(ctx, module, fh)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *Service) uploadSingle(ctx context.Context, module string, fileHeader *multipart.FileHeader) (domain.UploadResult, error) {
	if fileHeader == nil {
		return domain.UploadResult{}, domain.ErrEmptyFile
	}
	if fileHeader.Size == 0 {
		return domain.UploadResult{}, domain.ErrEmptyFile
	}

	file, err := fileHeader.Open()
	if err != nil {
		return domain.UploadResult{}, err
	}
	defer file.Close()

	contentType, err := detectContentType(file)
	if err != nil {
		return domain.UploadResult{}, err
	}

	category, ext, err := classifyFile(contentType, fileHeader.Filename)
	if err != nil {
		return domain.UploadResult{}, err
	}

	basePath := s.resolveBasePath(module, category)
	now := time.Now().UTC()
	datePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	dir := path.Join(basePath, datePath)

	filename := xid.New().String() + ext
	objectName := path.Join(dir, filename)

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return domain.UploadResult{}, err
	}

	if err := s.repo.PutObject(ctx, s.Bucket, objectName, file, fileHeader.Size, contentType); err != nil {
		return domain.UploadResult{}, err
	}

	return domain.UploadResult{
		Directory:        dir,
		Filename:         filename,
		OriginalFilename: fileHeader.Filename,
		MimeType:         contentType,
		Size:             fileHeader.Size,
		Category:         category,
	}, nil
}

func (s *Service) resolveBasePath(module, category string) string {
	root := s.basePath
	if module != "" {
		if modulePath, ok := s.modulePaths[module]; ok && modulePath != "" {
			if root != "" {
				return path.Join(root, modulePath)
			}
			return modulePath
		}
		if root != "" {
			return path.Join(root, module, category)
		}
		return path.Join(module, category)
	}

	if root != "" {
		return path.Join(root, category)
	}
	return category
}

func detectContentType(file multipart.File) (string, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	mime := http.DetectContentType(buf[:n])
	if mime == "" {
		mime = "application/octet-stream"
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	return mime, nil
}

func classifyFile(mimeType, originalFilename string) (string, string, error) {
	ext := strings.ToLower(filepath.Ext(originalFilename))
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		if ext == "" {
			ext = extensionFromMime(mimeType)
		}
		if ext == "" {
			ext = ".img"
		}
		return "images", ext, nil
	case strings.HasPrefix(mimeType, "video/"):
		if ext == "" {
			ext = extensionFromMime(mimeType)
		}
		if isAllowedVideoExt(ext) {
			return "videos", ext, nil
		}
		return "", "", domain.ErrUnsupportedType
	case mimeType == "application/pdf":
		return "documents", ".pdf", nil
	case mimeType == "application/zip" || mimeType == "application/x-zip-compressed":
		return "archives", ".zip", nil
	case mimeType == "application/x-rar-compressed" || mimeType == "application/vnd.rar":
		return "archives", ".rar", nil
	}

	switch ext {
	case ".pdf":
		return "documents", ".pdf", nil
	case ".zip":
		return "archives", ".zip", nil
	case ".rar":
		return "archives", ".rar", nil
	case ".mp4", ".mov", ".avi", ".mkv", ".webm":
		return "videos", ext, nil
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return "images", ext, nil
	default:
		return "", "", domain.ErrUnsupportedType
	}
}

func extensionFromMime(mime string) string {
	switch strings.ToLower(mime) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "video/mp4":
		return ".mp4"
	case "video/quicktime":
		return ".mov"
	case "video/x-msvideo":
		return ".avi"
	case "video/x-matroska":
		return ".mkv"
	case "video/webm":
		return ".webm"
	default:
		return ""
	}
}

func isAllowedVideoExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".mp4", ".mov", ".avi", ".mkv", ".webm":
		return true
	default:
		return false
	}
}
