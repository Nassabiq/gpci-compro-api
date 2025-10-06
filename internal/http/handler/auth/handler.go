package auth

import (
	"time"

	"github.com/Nassabiq/gpci-compro-api/internal/config"
	internalhandler "github.com/Nassabiq/gpci-compro-api/internal/http/handler/internal"
	"github.com/Nassabiq/gpci-compro-api/internal/http/response"
	usersservice "github.com/Nassabiq/gpci-compro-api/internal/modules/users/service"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	cfg   *config.Config
	users *usersservice.Service
}

type registerReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func New(cfg *config.Config, users *usersservice.Service) *Handler {
	return &Handler{cfg: cfg, users: users}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var in registerReq
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}
	if in.Name == "" || in.Email == "" || in.Password == "" {
		return response.Error(c, fiber.StatusBadRequest, "missing_fields", "name, email, and password are required", nil)
	}
	if u, _ := h.users.FindByEmail(internalhandler.ContextOrBackground(c), in.Email); u != nil {
		return response.Error(c, fiber.StatusConflict, "email_exists", "email already used", nil)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "hash_failed", err.Error(), nil)
	}
	id, err := h.users.Create(internalhandler.ContextOrBackground(c), in.Name, in.Email, string(hash))
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "user_create_failed", err.Error(), nil)
	}
	data := fiber.Map{"id": id, "email": in.Email}
	return response.Created(c, data)
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var in loginReq
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "invalid_body", "invalid request body", nil)
	}

	user, err := h.users.FindByEmail(internalhandler.ContextOrBackground(c), in.Email)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "user_lookup_failed", err.Error(), nil)
	}

	if user == nil || !user.IsActive {
		return response.Error(c, fiber.StatusUnauthorized, "invalid_credentials", "invalid credentials", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "invalid_credentials", "invalid credentials", nil)
	}

	claims := jwt.MapClaims{
		"sub":   user.XID,
		"email": user.Email,
		"exp":   time.Now().Add(h.cfg.Auth.JWTExpires).Unix(),
		"iat":   time.Now().Unix(),
		"iss":   h.cfg.App.Name,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(h.cfg.Auth.JWTSecret))

	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "token_sign_failed", err.Error(), nil)
	}

	data := fiber.Map{
		"access_token": signed,
		"token_type":   "Bearer",
		"expires_in":   int(h.cfg.Auth.JWTExpires.Seconds()),
	}

	return response.Success(c, fiber.StatusOK, data, nil)
}
