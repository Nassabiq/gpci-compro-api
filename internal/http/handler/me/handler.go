package me

import (
	"context"
	"database/sql"

	internalhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/internal"
	"github.com/Nassabiq/gpci-compro-api/internal/http/response"
	rbacservice "github.com/Nassabiq/gpci-compro-api/internal/modules/rbac/service"
	"github.com/Nassabiq/gpci-compro-api/internal/modules/users/domain"
	usersservice "github.com/Nassabiq/gpci-compro-api/internal/modules/users/service"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	Service *usersservice.Service
	RBAC    *rbacservice.Service
}

func New(service *usersservice.Service, rbac *rbacservice.Service) *Handler {
	return &Handler{Service: service, RBAC: rbac}
}

func (h *Handler) Profile(c *fiber.Ctx) error {
	user, err := h.currentUser(c)
	if err != nil {
		return err
	}
	ctx := internalhandler.ContextOrBackground(c)
	payload, err := h.profilePayload(ctx, user)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "profile_load_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, payload, nil)
}

func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	user, err := h.currentUser(c)
	if err != nil {
		return err
	}

	var payload domain.UserProfileUpdatePayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if err := internalhandler.ValidatePayload(c, &payload); err != nil {
		return err
	}

	ctx := internalhandler.ContextOrBackground(c)
	if payload.Email != user.Email {
		existing, err := h.Service.FindByEmail(ctx, payload.Email)
		if err != nil {
			return response.Error(c, fiber.StatusInternalServerError, "user_lookup_failed", err.Error(), nil)
		}
		if existing != nil && existing.XID != user.XID {
			return response.Error(c, fiber.StatusConflict, "email_exists", "email already in use", nil)
		}
	}

	updated, err := h.Service.Update(ctx, user.XID, payload.Name, payload.Email, nil, user.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "user_not_found", "user not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "user_update_failed", err.Error(), nil)
	}
	payloadResp, err := h.profilePayload(ctx, updated)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "profile_load_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, payloadResp, nil)
}

func (h *Handler) UpdatePassword(c *fiber.Ctx) error {
	user, err := h.currentUser(c)
	if err != nil {
		return err
	}

	var payload domain.UserPasswordUpdatePayload
	if err := c.BodyParser(&payload); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if err := internalhandler.ValidatePayload(c, &payload); err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.CurrentPassword)); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_current_password", "current password is incorrect", nil)
	}
	if payload.CurrentPassword == payload.Password {
		return response.Error(c, fiber.StatusBadRequest, "password_unchanged", "new password must be different from current password", nil)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "hash_failed", err.Error(), nil)
	}
	hashed := string(hash)

	ctx := internalhandler.ContextOrBackground(c)
	updated, err := h.Service.Update(ctx, user.XID, user.Name, user.Email, &hashed, user.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, fiber.StatusNotFound, "user_not_found", "user not found", nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, "user_update_failed", err.Error(), nil)
	}
	payloadResp, err := h.profilePayload(ctx, updated)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "profile_load_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, payloadResp, nil)
}

func (h *Handler) currentUser(c *fiber.Ctx) (*domain.User, error) {
	xid, _ := c.Locals("user_xid").(string)
	if xid == "" {
		return nil, response.Error(c, fiber.StatusUnauthorized, "unauthorized", "unauthorized", nil)
	}

	user, err := h.Service.FindByXID(internalhandler.ContextOrBackground(c), xid)
	if err != nil {
		return nil, response.Error(c, fiber.StatusInternalServerError, "user_lookup_failed", err.Error(), nil)
	}
	if user == nil {
		return nil, response.Error(c, fiber.StatusUnauthorized, "user_not_found", "user not found", nil)
	}
	if !user.IsActive {
		return nil, response.Error(c, fiber.StatusForbidden, "user_inactive", "user is inactive", nil)
	}

	return user, nil
}

func (h *Handler) profilePayload(ctx context.Context, user *domain.User) (fiber.Map, error) {
	roles, err := h.RBAC.ListUserRoles(ctx, user.XID)
	if err != nil {
		return nil, err
	}
	perms, err := h.RBAC.ListUserPermissions(ctx, user.XID)
	if err != nil {
		return nil, err
	}
	return fiber.Map{
		"xid":              user.XID,
		"name":             user.Name,
		"email":            user.Email,
		"is_active":        user.IsActive,
		"created_at":       user.CreatedAt,
		"updated_at":       user.UpdatedAt,
		"email_verified_at": user.EmailVerifiedAt,
		"roles":            roles,
		"permissions":      perms,
	}, nil
}
