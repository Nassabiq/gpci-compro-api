package faq

import (
	"database/sql"
	"strconv"

	internalhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/internal"
	"github.com/Nassabiq/gpci-compro-api/internal/http/response"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/faq/domain"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/faq/service"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
	filter := domain.FAQFilter{}
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

	faqs, err := h.Service.ListFAQs(internalhandler.ContextOrBackground(c), filter)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "faq_list_failed", err.Error(), nil)
	}
	meta := fiber.Map{"total": faqs.Total, "page": faqs.Page, "page_size": faqs.PageSize}
	return response.Success(c, fiber.StatusOK, faqs.Items, meta)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var payload domain.FAQPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if err := internalhandler.ValidatePayload(c, &payload); err != nil {
		return err
	}

	faq, err := h.Service.CreateFAQ(internalhandler.ContextOrBackground(c), payload)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "faq_create_failed", err.Error(), nil)
	}
	return response.Created(c, faq)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	id, err := parseFAQID(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_faq_id", "invalid FAQ id", nil)
	}

	faq, err := h.Service.GetFAQByID(internalhandler.ContextOrBackground(c), id)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "faq_not_found", "faq not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "faq_lookup_failed", err.Error(), nil)
	}
	if faq == nil {
		return response.Error(c, fiber.StatusNotFound, "faq_not_found", "faq not found", nil)
	}
	return response.Success(c, fiber.StatusOK, faq, nil)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := parseFAQID(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_faq_id", "invalid FAQ id", nil)
	}

	var payload domain.FAQPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if err := internalhandler.ValidatePayload(c, &payload); err != nil {
		return err
	}

	faq, err := h.Service.UpdateFAQ(internalhandler.ContextOrBackground(c), id, payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "faq_not_found", "faq not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "faq_update_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, faq, nil)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := parseFAQID(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_faq_id", "invalid FAQ id", nil)
	}

	if err := h.Service.DeleteFAQ(internalhandler.ContextOrBackground(c), id); err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "faq_not_found", "faq not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "faq_delete_failed", err.Error(), nil)
	}
	return response.NoContent(c)
}

func parseFAQID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}
