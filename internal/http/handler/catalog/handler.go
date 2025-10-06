package catalog

import (
	"database/sql"
	"strconv"

	internalhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/internal"
	"github.com/Nassabiq/gpci-compro-api/internal/http/response"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/catalog/domain"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/catalog/service"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) ListPrograms(c *fiber.Ctx) error {
	ctx := internalhandler.ContextOrBackground(c)
	items, err := h.Service.ListPrograms(ctx)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "program_list_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, items, nil)
}

func (h *Handler) CreateProgram(c *fiber.Ctx) error {
	var payload domain.ProgramPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if payload.ID == 0 || payload.Code == "" || payload.Name == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "id, code, and name are required", nil)
	}
	ctx := internalhandler.ContextOrBackground(c)
	program, err := h.Service.CreateProgram(ctx, payload)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "program_create_failed", err.Error(), nil)
	}
	return response.Created(c, program)
}

func (h *Handler) UpdateProgram(c *fiber.Ctx) error {
	idVal, err := strconv.ParseInt(c.Params("id"), 10, 16)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_program_id", "invalid program id", nil)
	}

	var payload domain.ProgramPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if payload.Code == "" || payload.Name == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "code and name are required", nil)
	}

	ctx := internalhandler.ContextOrBackground(c)
	program, err := h.Service.UpdateProgram(ctx, int16(idVal), payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "program_not_found", "program not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "program_update_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, program, nil)
}

func (h *Handler) DeleteProgram(c *fiber.Ctx) error {
	idVal, err := strconv.ParseInt(c.Params("id"), 10, 16)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_program_id", "invalid program id", nil)
	}
	ctx := internalhandler.ContextOrBackground(c)
	if err := h.Service.DeleteProgram(ctx, int16(idVal)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "program_delete_failed", err.Error(), nil)
	}
	return response.NoContent(c)
}

func (h *Handler) ListStatuses(c *fiber.Ctx) error {
	ctx := internalhandler.ContextOrBackground(c)
	items, err := h.Service.ListStatuses(ctx)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "status_list_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, items, nil)
}

func (h *Handler) CreateStatus(c *fiber.Ctx) error {
	var payload domain.StatusPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if payload.ID == 0 || payload.Code == "" || payload.Name == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "id, code, and name are required", nil)
	}
	ctx := internalhandler.ContextOrBackground(c)
	status, err := h.Service.CreateStatus(ctx, payload)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "status_create_failed", err.Error(), nil)
	}
	return response.Created(c, status)
}

func (h *Handler) UpdateStatus(c *fiber.Ctx) error {
	idVal, err := strconv.ParseInt(c.Params("id"), 10, 16)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_status_id", "invalid status id", nil)
	}

	var payload domain.StatusPayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if payload.Code == "" || payload.Name == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "code and name are required", nil)
	}

	ctx := internalhandler.ContextOrBackground(c)
	status, err := h.Service.UpdateStatus(ctx, int16(idVal), payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "status_not_found", "status not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "status_update_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, status, nil)
}

func (h *Handler) DeleteStatus(c *fiber.Ctx) error {
	idVal, err := strconv.ParseInt(c.Params("id"), 10, 16)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_status_id", "invalid status id", nil)
	}
	ctx := internalhandler.ContextOrBackground(c)
	if err := h.Service.DeleteStatus(ctx, int16(idVal)); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "status_delete_failed", err.Error(), nil)
	}
	return response.NoContent(c)
}
