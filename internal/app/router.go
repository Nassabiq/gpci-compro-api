package app

import (
	"errors"
	"net/http"

	authhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/auth"
	brandhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/brand"
	"github.com/Nassabiq/gpci-compro-api/internal/http/handler/catalog"
	faqhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/faq"
	"github.com/Nassabiq/gpci-compro-api/internal/http/handler/health"
	mehandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/me"
	producthandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/product"
	rbachandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/rbac"
	uploadshandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/uploads"
	userhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/user"
	"github.com/Nassabiq/gpci-compro-api/internal/http/middleware"
	"github.com/Nassabiq/gpci-compro-api/internal/http/response"
	brandmodule "github.com/Nassabiq/gpci-compro-api/internal/modules/brand"
	catalogmodule "github.com/Nassabiq/gpci-compro-api/internal/modules/catalog"
	faqmodule "github.com/Nassabiq/gpci-compro-api/internal/modules/faq"
	productmodule "github.com/Nassabiq/gpci-compro-api/internal/modules/product"
	rbacmodule "github.com/Nassabiq/gpci-compro-api/internal/modules/rbac"
	uploadsmodule "github.com/Nassabiq/gpci-compro-api/internal/modules/uploads"
	usersmodule "github.com/Nassabiq/gpci-compro-api/internal/modules/users"
	"github.com/Nassabiq/gpci-compro-api/internal/queue"
	"github.com/gofiber/fiber/v2"
	fibercors "github.com/gofiber/fiber/v2/middleware/cors"
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
	corsCfg := cfg.App.CORS
	corsConfig := fibercors.Config{
		AllowOrigins:     corsCfg.AllowOrigins,
		AllowMethods:     corsCfg.AllowMethods,
		AllowHeaders:     corsCfg.AllowHeaders,
		AllowCredentials: corsCfg.AllowCredentials,
		MaxAge:           corsCfg.MaxAge,
	}
	if corsCfg.ExposeHeaders != "" {
		corsConfig.ExposeHeaders = corsCfg.ExposeHeaders
	}
	app.Use(fibercors.New(corsConfig))

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
	gliCertHandler := producthandler.NewProgramCertificateHandler(productMod.ProgramCertService, "green_label")
	gtriCertHandler := producthandler.NewProgramCertificateHandler(productMod.ProgramCertService, "green_toll")
	faqMod := faqmodule.Provide(container.DB)
	faqHandler := faqhandler.New(faqMod.Service)
	userHandler := userhandler.New(usersMod.Service)
	rbacMod := rbacmodule.Provide(container.DB)
	rbacHandler := &rbachandler.RBACHandler{Service: rbacMod.Service}
	meHandler := mehandler.New(usersMod.Service, rbacMod.Service)

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

	authenticated := api.Group("", middleware.JWTAuth(jwtCfg))
	authenticated.Get("/profile", meHandler.Profile)
	authenticated.Put("/profile", meHandler.UpdateProfile)
	authenticated.Put("/profile/password", meHandler.UpdatePassword)

	meGroup := authenticated.Group("/me")
	meGroup.Get("", meHandler.Profile)
	meGroup.Put("/profile", meHandler.UpdateProfile)
	meGroup.Put("/password", meHandler.UpdatePassword)

	catalogGroup := authenticated.Group("/catalog")
	catalogGroup.Get("/programs", middleware.RequirePermission(rbacMod.Service, "catalog.programs.read"), catalogHandler.ListPrograms)
	catalogGroup.Post("/programs", middleware.RequirePermission(rbacMod.Service, "catalog.programs.write"), catalogHandler.CreateProgram)
	catalogGroup.Put("/programs/:id", middleware.RequirePermission(rbacMod.Service, "catalog.programs.write"), catalogHandler.UpdateProgram)
	catalogGroup.Delete("/programs/:id", middleware.RequirePermission(rbacMod.Service, "catalog.programs.delete"), catalogHandler.DeleteProgram)
	catalogGroup.Get("/statuses", middleware.RequirePermission(rbacMod.Service, "catalog.statuses.read"), catalogHandler.ListStatuses)
	catalogGroup.Post("/statuses", middleware.RequirePermission(rbacMod.Service, "catalog.statuses.write"), catalogHandler.CreateStatus)
	catalogGroup.Put("/statuses/:id", middleware.RequirePermission(rbacMod.Service, "catalog.statuses.write"), catalogHandler.UpdateStatus)
	catalogGroup.Delete("/statuses/:id", middleware.RequirePermission(rbacMod.Service, "catalog.statuses.delete"), catalogHandler.DeleteStatus)

	productGroup := authenticated.Group("/products")
	productGroup.Get("", middleware.RequirePermission(rbacMod.Service, "products.read"), productHandler.List)
	productGroup.Post("", middleware.RequirePermission(rbacMod.Service, "products.write"), productHandler.Create)
	productGroup.Get(":slug", middleware.RequirePermission(rbacMod.Service, "products.read"), productHandler.Get)
	productGroup.Put(":slug", middleware.RequirePermission(rbacMod.Service, "products.write"), productHandler.Update)
	productGroup.Delete(":slug", middleware.RequirePermission(rbacMod.Service, "products.delete"), productHandler.Delete)

	productGroup.Get(":slug/certifications", middleware.RequirePermission(rbacMod.Service, "product.certifications.read"), productCertHandler.ListProductCertifications)
	productGroup.Post(":slug/certifications", middleware.RequirePermission(rbacMod.Service, "product.certifications.write"), productCertHandler.CreateProductCertification)
	productGroup.Put(":slug/certifications/:certID", middleware.RequirePermission(rbacMod.Service, "product.certifications.write"), productCertHandler.UpdateProductCertification)
	productGroup.Delete(":slug/certifications/:certID", middleware.RequirePermission(rbacMod.Service, "product.certifications.delete"), productCertHandler.DeleteProductCertification)

	brandGroup := authenticated.Group("/brands")
	brandGroup.Get("", middleware.RequirePermission(rbacMod.Service, "brands.read"), brandHandler.ListBrands)
	brandGroup.Post("", middleware.RequirePermission(rbacMod.Service, "brands.write"), brandHandler.CreateBrand)
	brandGroup.Put(":id", middleware.RequirePermission(rbacMod.Service, "brands.write"), brandHandler.UpdateBrand)
	brandGroup.Delete(":id", middleware.RequirePermission(rbacMod.Service, "brands.delete"), brandHandler.DeleteBrand)

	categoryGroup := authenticated.Group("/brand-categories")
	categoryGroup.Get("", middleware.RequirePermission(rbacMod.Service, "brand.categories.read"), brandHandler.ListBrandCategories)
	categoryGroup.Post("", middleware.RequirePermission(rbacMod.Service, "brand.categories.write"), brandHandler.CreateBrandCategory)
	categoryGroup.Put(":id", middleware.RequirePermission(rbacMod.Service, "brand.categories.write"), brandHandler.UpdateBrandCategory)
	categoryGroup.Delete(":id", middleware.RequirePermission(rbacMod.Service, "brand.categories.delete"), brandHandler.DeleteBrandCategory)

	uploadsGroup := authenticated.Group("/uploads")
	uploadsGroup.Post("", middleware.RequirePermission(rbacMod.Service, "uploads.create"), uploadHandler.Upload)
	uploadsGroup.Post("/images", middleware.RequirePermission(rbacMod.Service, "uploads.create"), uploadHandler.Upload)

	faqGroup := authenticated.Group("/faqs")
	faqGroup.Get("", middleware.RequirePermission(rbacMod.Service, "faq.read"), faqHandler.List)
	faqGroup.Post("", middleware.RequirePermission(rbacMod.Service, "faq.write"), faqHandler.Create)
	faqGroup.Get("/:id", middleware.RequirePermission(rbacMod.Service, "faq.read"), faqHandler.Get)
	faqGroup.Put("/:id", middleware.RequirePermission(rbacMod.Service, "faq.write"), faqHandler.Update)
	faqGroup.Delete("/:id", middleware.RequirePermission(rbacMod.Service, "faq.delete"), faqHandler.Delete)

	gliGroup := authenticated.Group("/gli-certificates")
	gliGroup.Get("", middleware.RequirePermission(rbacMod.Service, "product.certifications.read"), gliCertHandler.List)
	gliGroup.Post("", middleware.RequirePermission(rbacMod.Service, "product.certifications.write"), gliCertHandler.Create)
	gliGroup.Put("/:slug/:certID", middleware.RequirePermission(rbacMod.Service, "product.certifications.write"), gliCertHandler.Update)
	gliGroup.Delete("/:slug/:certID", middleware.RequirePermission(rbacMod.Service, "product.certifications.delete"), gliCertHandler.Delete)

	gtriGroup := authenticated.Group("/gtri-certificates")
	gtriGroup.Get("", middleware.RequirePermission(rbacMod.Service, "product.certifications.read"), gtriCertHandler.List)
	gtriGroup.Post("", middleware.RequirePermission(rbacMod.Service, "product.certifications.write"), gtriCertHandler.Create)
	gtriGroup.Put("/:slug/:certID", middleware.RequirePermission(rbacMod.Service, "product.certifications.write"), gtriCertHandler.Update)
	gtriGroup.Delete("/:slug/:certID", middleware.RequirePermission(rbacMod.Service, "product.certifications.delete"), gtriCertHandler.Delete)

	rbacGroup := authenticated.Group("/rbac")
	rbacGroup.Post("/roles", middleware.RequirePermission(rbacMod.Service, "rbac.roles.write"), rbacHandler.CreateRole)
	rbacGroup.Post("/permissions", middleware.RequirePermission(rbacMod.Service, "rbac.permissions.write"), rbacHandler.CreatePermission)
	rbacGroup.Get("/roles", middleware.RequirePermission(rbacMod.Service, "rbac.roles.read"), rbacHandler.ListRoles)
	rbacGroup.Get("/permissions", middleware.RequirePermission(rbacMod.Service, "rbac.permissions.read"), rbacHandler.ListPermissions)
	rbacGroup.Post("/roles/:role/permissions", middleware.RequirePermission(rbacMod.Service, "rbac.roles.assign"), rbacHandler.AssignPermissionToRole)
	rbacGroup.Post("/users/:xid/roles", middleware.RequirePermission(rbacMod.Service, "rbac.users.assign_role"), rbacHandler.AssignRoleToUser)

	usersGroup := api.Group("/users", middleware.JWTAuth(jwtCfg))
	usersGroup.Get("", middleware.RequirePermission(rbacMod.Service, "users.read"), userHandler.List)
	usersGroup.Get("/:xid", middleware.RequirePermission(rbacMod.Service, "users.read"), userHandler.Get)
	usersGroup.Post("", middleware.RequirePermission(rbacMod.Service, "users.write"), userHandler.Create)
	usersGroup.Put("/:xid", middleware.RequirePermission(rbacMod.Service, "users.write"), userHandler.Update)
	usersGroup.Delete("/:xid", middleware.RequirePermission(rbacMod.Service, "users.delete"), userHandler.Delete)

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
