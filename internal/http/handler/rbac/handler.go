package rbac

import (
	internalhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/internal"
	"github.com/Nassabiq/gpci-compro-api/internal/http/response"
	rbacservice "github.com/Nassabiq/gpci-compro-api/internal/modules/rbac/service"
	"github.com/gofiber/fiber/v2"
)

type RBACHandler struct {
	Service *rbacservice.Service
}

func (h *RBACHandler) CreateRole(c *fiber.Ctx) error {
	var in struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&in); err != nil || in.Name == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "name required", nil)
	}
	id, err := h.Service.CreateRole(internalhandler.ContextOrBackground(c), in.Name)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "role_create_failed", err.Error(), nil)
	}
	return response.Created(c, fiber.Map{"id": id, "name": in.Name})
}

func (h *RBACHandler) CreatePermission(c *fiber.Ctx) error {
	var in struct {
		Key         string `json:"key"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&in); err != nil || in.Key == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "key required", nil)
	}
	id, err := h.Service.CreatePermission(internalhandler.ContextOrBackground(c), in.Key, in.Description)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "permission_create_failed", err.Error(), nil)
	}
	return response.Created(c, fiber.Map{"id": id, "key": in.Key})
}

func (h *RBACHandler) ListRoles(c *fiber.Ctx) error {
	roles, err := h.Service.ListRoles(internalhandler.ContextOrBackground(c))
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "role_list_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, roles, nil)
}

func (h *RBACHandler) ListPermissions(c *fiber.Ctx) error {
	perms, err := h.Service.ListPermissions(internalhandler.ContextOrBackground(c))
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "permission_list_failed", err.Error(), nil)
	}
	return response.Success(c, fiber.StatusOK, perms, nil)
}

func (h *RBACHandler) AssignPermissionToRole(c *fiber.Ctx) error {
	role := c.Params("role")
	var in struct {
		Permission string `json:"permission"`
	}
	if role == "" || c.BodyParser(&in) != nil || in.Permission == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "role and permission required", nil)
	}
	if err := h.Service.AssignPermissionToRole(internalhandler.ContextOrBackground(c), role, in.Permission); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "assign_permission_failed", err.Error(), nil)
	}
	return response.NoContent(c)
}

func (h *RBACHandler) AssignRoleToUser(c *fiber.Ctx) error {
	userXID := c.Params("xid")
	var in struct {
		Role string `json:"role"`
	}
	if userXID == "" || c.BodyParser(&in) != nil || in.Role == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "user xid and role required", nil)
	}
	if err := h.Service.AssignRoleToUserByXID(internalhandler.ContextOrBackground(c), userXID, in.Role); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "assign_role_failed", err.Error(), nil)
	}
	return response.NoContent(c)
}
