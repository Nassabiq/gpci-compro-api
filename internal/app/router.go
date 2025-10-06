package app

import (
	"errors"
	"net/http"

	authhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/auth"
	brandhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/brand"
	"github.com/Nassabiq/gpci-compro-api/internal/http/handler/catalog"
	"github.com/Nassabiq/gpci-compro-api/internal/http/handler/health"
	producthandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/product"
	rbachandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/rbac"
	uploadshandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/uploads"
	"github.com/Nassabiq/gpci-compro-api/internal/http/middleware"
	"github.com/Nassabiq/gpci-compro-api/internal/http/response"
	brandmodule "github.com/Nassabiq/gpci-compro-api/internal/modules/brand"
	catalogmodule "github.com/Nassabiq/gpci-compro-api/internal/modules/catalog"
	productmodule "github.com/Nassabiq/gpci-compro-api/internal/modules/product"
	rbacmodule "github.com/Nassabiq/gpci-compro-api/internal/modules/rbac"
	uploadsmodule "github.com/Nassabiq/gpci-compro-api/internal/modules/uploads"
	usersmodule "github.com/Nassabiq/gpci-compro-api/internal/modules/users"
	"github.com/Nassabiq/gpci-compro-api/internal/queue"
	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
)

func Setup(container *Container) *fiber.App {
	cfg := container.Config

	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
		IdleTimeout:  cfg.App.IdleTimeout,
		ErrorHandler: errorHandler,
	})

	app.Use(middleware.RequestID())
	app.Use(fiberlogger.New())

	usersMod := usersmodule.Provide(container.DB)
	auth := authhandler.New(cfg, usersMod.Service)

	catalogMod := catalogmodule.Provide(container.DB)
	catalogHandler := catalog.New(catalogMod.Service)

	brandMod := brandmodule.Provide(container.DB)
	brandHandler := brandhandler.New(brandMod.Service)

	productMod := productmodule.Provide(container.DB)
	productHandler := producthandler.New(productMod.Service)
	productCertHandler := &producthandler.ProductCertificationHandler{
		Service:        productMod.CertificationService,
		ProductService: productMod.Service,
	}

	rbacMod := rbacmodule.Provide(container.DB)
	rbacHandler := &rbachandler.RBACHandler{Service: rbacMod.Service}

	uploadsMod := uploadsmodule.Provide(container.Storage, container.Config.Storage.Bucket, container.Config.Storage.BasePath, uploadsModulePaths())
	uploadHandler := &uploadshandler.UploadHandler{Service: uploadsMod.Service}

	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	api := app.Group("/api")

	api.Get("/ping", func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, fiber.Map{"pong": true}, nil)
	})

	api.Post("/notify", func(c *fiber.Ctx) error {
		var req struct {
			UserID  int64  `json:"user_id"`
			Message string `json:"message"`
		}
		if err := c.BodyParser(&req); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
		}
		if req.UserID == 0 || req.Message == "" {
			return response.Error(c, fiber.StatusBadRequest, "missing_fields", "user_id and message required", nil)
		}
		if _, err := queue.EnqueueNotifyUser(container.AsynqClient, queue.NotifyUserPayload{UserID: req.UserID, Message: req.Message}); err != nil {
			return response.Error(c, http.StatusInternalServerError, "notify_enqueue_failed", err.Error(), nil)
		}
		return response.Success(c, http.StatusAccepted, fiber.Map{"queued": true}, nil)
	})

	api.Post("/auth/register", auth.Register)
	api.Post("/auth/login", auth.Login)

	api.Get("/health", health.Check)

	jwtCfg := middleware.JWTConfig{Secret: cfg.Auth.JWTSecret}

	meGroup := api.Group("/me", middleware.JWTAuth(jwtCfg))
	meGroup.Get("", func(c *fiber.Ctx) error {
		return response.Success(c, fiber.StatusOK, fiber.Map{"user_xid": c.Locals("user_xid")}, nil)
	})

	catalogGroup := api.Group("/catalog")
	catalogGroup.Get("/programs", catalogHandler.ListPrograms)
	catalogGroup.Post("/programs", catalogHandler.CreateProgram)
	catalogGroup.Put("/programs/:id", catalogHandler.UpdateProgram)
	catalogGroup.Delete("/programs/:id", catalogHandler.DeleteProgram)
	catalogGroup.Get("/statuses", catalogHandler.ListStatuses)
	catalogGroup.Post("/statuses", catalogHandler.CreateStatus)
	catalogGroup.Put("/statuses/:id", catalogHandler.UpdateStatus)
	catalogGroup.Delete("/statuses/:id", catalogHandler.DeleteStatus)

	brandGroup := catalogGroup.Group("/brands")
	brandGroup.Get("", brandHandler.ListBrands)
	brandGroup.Post("", brandHandler.CreateBrand)
	brandGroup.Put(":id", brandHandler.UpdateBrand)
	brandGroup.Delete(":id", brandHandler.DeleteBrand)

	categoryGroup := catalogGroup.Group("/brand-categories")
	categoryGroup.Get("", brandHandler.ListBrandCategories)
	categoryGroup.Post("", brandHandler.CreateBrandCategory)
	categoryGroup.Put(":id", brandHandler.UpdateBrandCategory)
	categoryGroup.Delete(":id", brandHandler.DeleteBrandCategory)

	productGroup := catalogGroup.Group("/products")
	productGroup.Get("", productHandler.List)
	productGroup.Post("", productHandler.Create)
	productGroup.Get(":slug", productHandler.Get)
	productGroup.Put(":slug", productHandler.Update)
	productGroup.Delete(":slug", productHandler.Delete)

	productGroup.Get(":slug/certifications", productCertHandler.ListProductCertifications)
	productGroup.Post(":slug/certifications", productCertHandler.CreateProductCertification)
	productGroup.Put(":slug/certifications/:certID", productCertHandler.UpdateProductCertification)
	productGroup.Delete(":slug/certifications/:certID", productCertHandler.DeleteProductCertification)

	uploadsGroup := api.Group("/uploads", middleware.JWTAuth(jwtCfg))
	uploadsGroup.Post("", uploadHandler.Upload)
	uploadsGroup.Post("/images", uploadHandler.Upload)

	rbacGroup := api.Group("/rbac", middleware.JWTAuth(jwtCfg))
	rbacGroup.Post("/roles", middleware.RequirePermission(rbacMod.Service, "rbac.roles.write"), rbacHandler.CreateRole)
	rbacGroup.Post("/permissions", middleware.RequirePermission(rbacMod.Service, "rbac.permissions.write"), rbacHandler.CreatePermission)
	rbacGroup.Get("/roles", middleware.RequirePermission(rbacMod.Service, "rbac.roles.read"), rbacHandler.ListRoles)
	rbacGroup.Get("/permissions", middleware.RequirePermission(rbacMod.Service, "rbac.permissions.read"), rbacHandler.ListPermissions)
	rbacGroup.Post("/roles/:role/permissions", middleware.RequirePermission(rbacMod.Service, "rbac.roles.assign"), rbacHandler.AssignPermissionToRole)
	rbacGroup.Post("/users/:xid/roles", middleware.RequirePermission(rbacMod.Service, "rbac.users.assign_role"), rbacHandler.AssignRoleToUser)

	return app
}

func errorHandler(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return response.Error(c, fiberErr.Code, "request_error", fiberErr.Message, nil)
	}

	return response.Error(c, http.StatusInternalServerError, "internal_error", err.Error(), nil)
}

func uploadsModulePaths() map[string]string {
	return map[string]string{
		"product":       "images/products",
		"certification": "images/certifications",
		"company":       "images/companies",
		"brand":         "images/brands",
		"document":      "documents",
	}
}
