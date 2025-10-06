package brand

import (
	"database/sql"
	"strconv"

	internalhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/internal"
	"github.com/Nassabiq/gpci-compro-api/internal/http/response"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/brand/domain"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/brand/service"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Service *service.BrandService
}

func New(service *service.BrandService) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) ListBrandCategories(c *fiber.Ctx) error {
	filter := domain.BrandCategoryFilter{}
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

	categories, err := h.Service.ListBrandCategories(internalhandler.ContextOrBackground(c), filter)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "brand_category_list_failed", err.Error(), nil)
	}
	meta := fiber.Map{"total": categories.Total, "page": categories.Page, "page_size": categories.PageSize}
	return response.Success(c, fiber.StatusOK, categories.Items, meta)
}

func (h *Handler) CreateBrandCategory(c *fiber.Ctx) error {
	var payload domain.BrandCategoryPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if payload.Name == "" || payload.Slug == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "name and slug are required", nil)
	}

	category, err := h.Service.CreateBrandCategory(internalhandler.ContextOrBackground(c), payload)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "brand_category_create_failed", err.Error(), nil)
	}
	return response.Created(c, category)
}

func (h *Handler) UpdateBrandCategory(c *fiber.Ctx) error {
	idVal, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_brand_category_id", "invalid brand category id", nil)
	}

	var payload domain.BrandCategoryPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if payload.Name == "" || payload.Slug == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "name and slug are required", nil)
	}

	category, err := h.Service.UpdateBrandCategory(internalhandler.ContextOrBackground(c), idVal, payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "brand_category_not_found", "brand category not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "brand_category_update_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, category, nil)
}

func (h *Handler) DeleteBrandCategory(c *fiber.Ctx) error {
	idVal, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_brand_category_id", "invalid brand category id", nil)
	}
	if err := h.Service.DeleteBrandCategory(internalhandler.ContextOrBackground(c), idVal); err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "brand_category_not_found", "brand category not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "brand_category_delete_failed", err.Error(), nil)
	}
	return response.NoContent(c)
}

func (h *Handler) ListBrands(c *fiber.Ctx) error {
	filter := domain.BrandFilter{CategorySlug: c.Query("category")}
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

	brands, err := h.Service.ListBrands(internalhandler.ContextOrBackground(c), filter)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "brand_list_failed", err.Error(), nil)
	}
	meta := fiber.Map{"total": brands.Total, "page": brands.Page, "page_size": brands.PageSize}
	return response.Success(c, fiber.StatusOK, brands.Items, meta)
}

func (h *Handler) CreateBrand(c *fiber.Ctx) error {
	var payload domain.BrandPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if payload.CategoryID == 0 || payload.Name == "" || payload.Slug == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "category_id, name, and slug are required", nil)
	}

	brand, err := h.Service.CreateBrand(internalhandler.ContextOrBackground(c), payload)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "brand_create_failed", err.Error(), nil)
	}
	return response.Created(c, brand)
}

func (h *Handler) UpdateBrand(c *fiber.Ctx) error {
	idVal, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_brand_id", "invalid brand id", nil)
	}

	var payload domain.BrandPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if payload.CategoryID == 0 || payload.Name == "" || payload.Slug == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "category_id, name, and slug are required", nil)
	}

	brand, err := h.Service.UpdateBrand(internalhandler.ContextOrBackground(c), idVal, payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "brand_not_found", "brand not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "brand_update_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, brand, nil)
}

func (h *Handler) DeleteBrand(c *fiber.Ctx) error {
	idVal, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_brand_id", "invalid brand id", nil)
	}
	if err := h.Service.DeleteBrand(internalhandler.ContextOrBackground(c), idVal); err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "brand_not_found", "brand not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "brand_delete_failed", err.Error(), nil)
	}
	return response.NoContent(c)
}
