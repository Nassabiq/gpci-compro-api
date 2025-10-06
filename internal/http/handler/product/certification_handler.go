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

type ProductCertificationHandler struct {
	Service        *service.ProductCertificationService
	ProductService *service.ProductService
}

func (h *ProductCertificationHandler) ListProductCertifications(c *fiber.Ctx) error {
	ctx := internalhandler.ContextOrBackground(c)
	product, err := h.ProductService.GetProductBySlug(ctx, c.Params("slug"))

	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "product_not_found", "product not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "product_lookup_failed", err.Error(), nil)
	}

	filter := domain.ProductCertificationFilter{}
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

	result, err := h.Service.ListProductCertifications(ctx, c.Params("slug"), filter)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "product_certification_list_failed", err.Error(), nil)
	}

	meta := fiber.Map{
		"total":     result.Total,
		"page":      result.Page,
		"page_size": result.PageSize,
	}

	payload := fiber.Map{
		"product":        product,
		"certifications": result.Items,
	}

	return response.Success(c, fiber.StatusOK, payload, meta)
}

func (h *ProductCertificationHandler) CreateProductCertification(c *fiber.Ctx) error {
	var payload domain.ProductCertificationPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}

	if payload.CertificationID == 0 {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "certification_id is required", nil)
	}

	cert, err := h.Service.CreateProductCertification(internalhandler.ContextOrBackground(c), c.Params("slug"), payload)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "product_certification_create_failed", err.Error(), nil)
	}
	return response.Created(c, cert)
}

func (h *ProductCertificationHandler) UpdateProductCertification(c *fiber.Ctx) error {
	certID, err := strconv.ParseInt(c.Params("certID"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_certification_id", "invalid certification id", nil)
	}

	var payload domain.ProductCertificationPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}

	cert, err := h.Service.UpdateProductCertification(internalhandler.ContextOrBackground(c), c.Params("slug"), certID, payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "product_certification_not_found", "product certification not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "product_certification_update_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, cert, nil)
}

func (h *ProductCertificationHandler) DeleteProductCertification(c *fiber.Ctx) error {
	certID, err := strconv.ParseInt(c.Params("certID"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_certification_id", "invalid certification id", nil)
	}

	if err := h.Service.DeleteProductCertification(internalhandler.ContextOrBackground(c), c.Params("slug"), certID); err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "product_certification_not_found", "product certification not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "product_certification_delete_failed", err.Error(), nil)
	}

	return response.NoContent(c)
}
