package handler

import (
	"github.com/KejarBahasa/kejarbill-api/internal/module/group/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/group/service"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/request"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/gofiber/fiber/v3"
)

type GroupHandler struct {
	groupService *service.GroupService
}

func NewGroupHandler(
	groupService *service.GroupService,
) *GroupHandler {

	return &GroupHandler{
		groupService: groupService,
	}
}

func (h *GroupHandler) Create(c fiber.Ctx) error {
	var body dto.CreateGroupBodyRequest
	if err := request.ValidateBody(c, &body); err != nil {
		return request.HandleValidationError(c, err)
	}

	userID := security.GetUserID(c)

	groupID, err := h.groupService.Create(c.Context(), userID, body.Name)
	if err != nil {
		return err
	}

	return response.Success(c, "group created", fiber.Map{
		"group_id": groupID,
	})
}
