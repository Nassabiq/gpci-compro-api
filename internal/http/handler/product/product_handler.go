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

type Handler struct {
	Service *service.ProductService
}

func New(service *service.ProductService) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
	ctx := internalhandler.ContextOrBackground(c)
	filter := domain.ProductFilter{
		ProgramCode:  c.Query("program"),
		BrandSlug:    c.Query("brand"),
		CategorySlug: c.Query("category"),
		Search:       c.Query("search"),
		IsActiveOnly: internalhandler.ParseBoolQuery(c.Query("is_active")),
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

	result, err := h.Service.ListProducts(ctx, filter)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "product_list_failed", err.Error(), nil)
	}

	meta := fiber.Map{
		"total":     result.Total,
		"page":      result.Page,
		"page_size": result.PageSize,
	}

	return response.Success(c, fiber.StatusOK, result.Items, meta)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var payload domain.ProductPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if payload.CompanyID == 0 || payload.BrandID == 0 || payload.ProgramID == 0 || payload.Name == "" || payload.Slug == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "company_id, brand_id, program_id, name, and slug are required", nil)
	}

	product, err := h.Service.CreateProduct(internalhandler.ContextOrBackground(c), payload)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "product_create_failed", err.Error(), nil)
	}
	return response.Created(c, product)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	product, err := h.Service.GetProductBySlug(internalhandler.ContextOrBackground(c), c.Params("slug"))
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "product_not_found", "product not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "product_lookup_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, product, nil)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var payload domain.ProductPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if payload.CompanyID == 0 || payload.BrandID == 0 || payload.ProgramID == 0 || payload.Name == "" || payload.Slug == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "company_id, brand_id, program_id, name, and slug are required", nil)
	}

	product, err := h.Service.UpdateProduct(internalhandler.ContextOrBackground(c), c.Params("slug"), payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "product_not_found", "product not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "product_update_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, product, nil)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.Service.DeleteProduct(internalhandler.ContextOrBackground(c), c.Params("slug")); err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "product_not_found", "product not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "product_delete_failed", err.Error(), nil)
	}
	return response.NoContent(c)
}
