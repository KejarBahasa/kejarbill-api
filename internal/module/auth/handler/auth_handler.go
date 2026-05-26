package handler

import (
	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/service"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/request"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/utils"

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
	if err := request.ValidateBody(c, &req); err != nil {
		return request.HandleValidationError(c, err)
	}

	err := h.authService.Register(c.Context(), req)

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return response.Success[any](c, "register success", nil)
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dto.LoginRequest
	if err := request.ValidateBody(c, &req); err != nil {
		return request.HandleValidationError(c, err)
	}

	clientInfo := utils.GetClientInfo(c)

	resp, err := h.authService.Login(c.Context(), req, clientInfo)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	if clientInfo.ClientType == constants.ClientTypeWeb {
		security.SetRefreshCookie(c, resp.RefreshToken)
		resp.RefreshToken = ""
	}

	return response.Success(c, "login success", resp)
}

func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")

	if refreshToken == "" {
		refreshToken = c.Get("X-Refresh-Token")
	}

	if refreshToken == "" {
		return fiber.ErrUnauthorized
	}

	clientInfo := utils.GetClientInfo(c)

	resp, err := h.authService.RefreshToken(c.Context(), refreshToken, clientInfo)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	if clientInfo.ClientType == constants.ClientTypeWeb {
		security.SetRefreshCookie(c, resp.RefreshToken)
		resp.RefreshToken = ""
	}

	return response.Success(c, "refresh token success", resp)
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")

	if refreshToken == "" {
		refreshToken = c.Get("X-Refresh-Token")
	}

	if refreshToken == "" {
		return fiber.ErrUnauthorized
	}

	err := h.authService.Logout(c.Context(), refreshToken)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	security.ClearRefreshCookie(c)

	return response.Success[any](c, "logout success", nil)
}
