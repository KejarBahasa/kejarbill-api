package handler

import (
	"github.com/KejarBahasa/kejarbill-api/internal/module/user/service"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(
	userService *service.UserService,
) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Me(c fiber.Ctx) error {
	userID := security.GetUserID(c)

	data, err := h.userService.Me(c.Context(), userID)
	if err != nil {
		return fiber.ErrNotFound
	}

	return response.Success(c, "ok", data)
}
