package product

import (
	"database/sql"
	"strconv"

	internalhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/internal"
	"github.com/Nassabiq/gpci-compro-api/internal/http/response"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/product/domain"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/product/service"
	"github.com/gofiber/fiber/v2"
)

type ProgramCertificateHandler struct {
	Service     *service.ProgramCertificateService
	ProgramCode string
}

func NewProgramCertificateHandler(service *service.ProgramCertificateService, programCode string) *ProgramCertificateHandler {
	return &ProgramCertificateHandler{
		Service:     service,
		ProgramCode: programCode,
	}
}

func (h *ProgramCertificateHandler) List(c *fiber.Ctx) error {
	filter := domain.ProgramCertificateFilter{
		Search: c.Query("search"),
	}
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			filter.Page = page
		}
	}
	if sizeStr := c.Query("page_size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil {
			filter.PageSize = size
		}
	}

	result, err := h.Service.List(internalhandler.ContextOrBackground(c), h.ProgramCode, filter)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "program_certificate_list_failed", err.Error(), nil)
	}

	meta := fiber.Map{
		"total":     result.Total,
		"page":      result.Page,
		"page_size": result.PageSize,
	}
	return response.Success(c, fiber.StatusOK, result.Items, meta)
}

func (h *ProgramCertificateHandler) Create(c *fiber.Ctx) error {
	var payload domain.ProgramCertificatePayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if err := internalhandler.ValidatePayload(c, &payload); err != nil {
		return err
	}
	if payload.IssueDate != nil && payload.ExpiryDate != nil && payload.IssueDate.After(*payload.ExpiryDate) {
		return response.Error(c, fiber.StatusBadRequest, "validation_failed", "expiry_date must be after issue_date", fiber.Map{"expiry_date": "must be after issue_date"})
	}

	record, err := h.Service.Create(internalhandler.ContextOrBackground(c), h.ProgramCode, payload)
	if err != nil {
		return h.handleServiceError(c, err, "program_certificate_create_failed")
	}

	return response.Created(c, record)
}

func (h *ProgramCertificateHandler) Update(c *fiber.Ctx) error {
	productSlug := c.Params("slug")
	certID, err := strconv.ParseInt(c.Params("certID"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_certification_id", "invalid certification id", nil)
	}

	var payload domain.ProductCertificationPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if err := internalhandler.ValidatePayload(c, &payload); err != nil {
		return err
	}
	if payload.IssueDate != nil && payload.ExpiryDate != nil && payload.IssueDate.After(*payload.ExpiryDate) {
		return response.Error(c, fiber.StatusBadRequest, "validation_failed", "expiry_date must be after issue_date", fiber.Map{"expiry_date": "must be after issue_date"})
	}

	record, err := h.Service.Update(internalhandler.ContextOrBackground(c), h.ProgramCode, productSlug, certID, payload)
	if err != nil {
		return h.handleServiceError(c, err, "program_certificate_update_failed")
	}
	return response.Success(c, fiber.StatusOK, record, nil)
}

func (h *ProgramCertificateHandler) Delete(c *fiber.Ctx) error {
	productSlug := c.Params("slug")
	certID, err := strconv.ParseInt(c.Params("certID"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_certification_id", "invalid certification id", nil)
	}

	if err := h.Service.Delete(internalhandler.ContextOrBackground(c), h.ProgramCode, productSlug, certID); err != nil {
		return h.handleServiceError(c, err, "program_certificate_delete_failed")
	}
	return response.NoContent(c)
}

func (h *ProgramCertificateHandler) handleServiceError(c *fiber.Ctx, err error, code string) error {
	switch {
	case err == nil:
		return nil
	case err == service.ErrProductNotFound:
		return response.Error(c, fiber.StatusNotFound, "product_not_found", "product not found", nil)
	case err == service.ErrCertificationNotFound:
		return response.Error(c, fiber.StatusNotFound, "certification_not_found", "certification not found", nil)
	case err == service.ErrProgramMismatch:
		return response.Error(c, fiber.StatusBadRequest, "program_mismatch", "product or certification does not match program", nil)
	case err == service.ErrCertificateNotFound:
		return response.Error(c, fiber.StatusNotFound, "program_certificate_not_found", "certificate not found", nil)
	case err == sql.ErrNoRows:
		return response.Error(c, fiber.StatusNotFound, "program_certificate_not_found", "certificate not found", nil)
	default:
		return response.Error(c, fiber.StatusInternalServerError, code, err.Error(), nil)
	}
}
