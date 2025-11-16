package user

import (
	"database/sql"
	"errors"

	internalhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/internal"
	"github.com/Nassabiq/gpci-compro-api/internal/http/response"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/users/domain"
	usersservice "github.com/Nassabiq/gpci-compro-api/internal/modules/users/service"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	Service *usersservice.Service
}

func New(service *usersservice.Service) *Handler {
	return &Handler{Service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
	users, err := h.Service.List(internalhandler.ContextOrBackground(c))
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "users_list_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, users, nil)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	user, err := h.Service.FindByXID(internalhandler.ContextOrBackground(c), c.Params("xid"))
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "users_lookup_failed", err.Error(), nil)
	}
	if user == nil {
		return response.Error(c, fiber.StatusNotFound, "user_not_found", "user not found", nil)
	}
	return response.Success(c, fiber.StatusOK, user, nil)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var payload domain.UserCreatePayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if err := internalhandler.ValidatePayload(c, &payload); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "hash_failed", err.Error(), nil)
	}

	ctx := internalhandler.ContextOrBackground(c)
	isActive := true
	if payload.IsActive != nil {
		isActive = *payload.IsActive
	}

	id, err := h.Service.Create(ctx, payload.Name, payload.Email, string(hash), isActive)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "user_create_failed", err.Error(), nil)
	}

	user, err := h.Service.FindByID(ctx, id)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "user_lookup_failed", err.Error(), nil)
	}
	if user == nil {
		return response.Error(c, fiber.StatusInternalServerError, "user_lookup_failed", "user not found after creation", nil)
	}

	return response.Created(c, user)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var payload domain.UserUpdatePayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if err := internalhandler.ValidatePayload(c, &payload); err != nil {
		return err
	}

	ctx := internalhandler.ContextOrBackground(c)
	existing, err := h.Service.FindByXID(ctx, c.Params("xid"))
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "users_lookup_failed", err.Error(), nil)
	}
	if existing == nil {
		return response.Error(c, fiber.StatusNotFound, "user_not_found", "user not found", nil)
	}

	isActive := existing.IsActive
	if payload.IsActive != nil {
		isActive = *payload.IsActive
	}

	var passwordHash *string
	if payload.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*payload.Password), bcrypt.DefaultCost)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "hash_failed", err.Error(), nil)
		}
		hashed := string(hash)
		passwordHash = &hashed
	}

	user, err := h.Service.Update(ctx, existing.XID, payload.Name, payload.Email, passwordHash, isActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.Error(c, fiber.StatusNotFound, "user_not_found", "user not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "user_update_failed", err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, user, nil)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.Service.Delete(internalhandler.ContextOrBackground(c), c.Params("xid")); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.Error(c, fiber.StatusNotFound, "user_not_found", "user not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "user_delete_failed", err.Error(), nil)
	}
	return response.NoContent(c)
}
