package uploads

import (
	"errors"
	"mime/multipart"
	"net/http"

	internalhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/internal"
	"github.com/Nassabiq/gpci-compro-api/internal/http/response"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/uploads/domain"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/uploads/service"
	"github.com/gofiber/fiber/v2"
)

type UploadHandler struct {
	Service *service.Service
}

func (h *UploadHandler) Upload(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	var fileHeaders []*multipart.FileHeader

	switch {
	case err == nil && form != nil:
		if files := form.File["files"]; len(files) > 0 {
			fileHeaders = append(fileHeaders, files...)
		}
		if singles := form.File["file"]; len(singles) > 0 {
			fileHeaders = append(fileHeaders, singles...)
		}
	default:
		if fh, ferr := c.FormFile("file"); ferr == nil && fh != nil {
			fileHeaders = append(fileHeaders, fh)
		} else {
			return fiber.NewError(fiber.StatusBadRequest, "file is required")
		}
	}

	if len(fileHeaders) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	}

	module := c.Query("module")
	if module == "" {
		module = c.FormValue("module")
	}

	results, err := h.Service.UploadFiles(internalhandler.ContextOrBackground(c), module, fileHeaders)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmptyFile),
			errors.Is(err, domain.ErrUnsupportedType),
			errors.Is(err, domain.ErrNoFilesProvided):
			return response.Error(c, fiber.StatusBadRequest, "invalid_file", err.Error(), nil)
		default:
			return response.Error(c, http.StatusInternalServerError, "upload_failed", err.Error(), nil)
		}
	}

	return response.Created(c, fiber.Map{"files": results})
}
