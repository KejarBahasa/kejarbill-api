package handler

import (
	"strings"
	"unicode/utf8"

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

func (h *UserHandler) Search(c fiber.Ctx) error {
	keyword := strings.TrimSpace(c.Query("q"))
	if utf8.RuneCountInString(keyword) < 2 {
		return fiber.NewError(fiber.StatusBadRequest, "q must be at least 2 characters")
	}

	requesterUserID := security.GetUserID(c)

	users, err := h.userService.Search(c.Context(), requesterUserID, keyword)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return response.Success(c, "ok", fiber.Map{"users": users})
}
