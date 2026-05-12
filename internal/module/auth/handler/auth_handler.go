package handler

import (
	"time"

	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/service"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(
	authService *service.AuthService,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req dto.RegisterRequest

	if err := c.Bind().Body(&req); err != nil {
		return fiber.ErrBadRequest
	}

	err := h.authService.Register(c.Context(), req)

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return response.Success(c, "register success", nil)
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dto.LoginRequest

	if err := c.Bind().Body(&req); err != nil {
		return fiber.ErrBadRequest
	}

	resp, err := h.authService.Login(c.Context(), req)

	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	return response.Success(c, "login success", resp)
}

func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	var req dto.RefreshTokenRequest

	if err := c.Bind().Body(&req); err != nil {
		return fiber.ErrBadRequest
	}

	resp, err := h.authService.RefreshToken(c.Context(), req)

	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	return response.Success(c, "refresh token success", resp)
}

func (h *AuthHandler) Me(c fiber.Ctx) error {
	userId := security.GetUserID(c)
	if userId == "" {
		return fiber.ErrUnauthorized
	}

	user, err := h.authService.Me(c.Context(), userId)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	resp := dto.MeResponse{
		ID:        user.ID,
		Name:      user.Name,
		Username:  user.Username,
		Email:     user.Email,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format(time.DateTime),
		UpdatedAt: user.UpdatedAt.Format(time.DateTime),
	}

	return response.Success(c, "ok", resp)
}
