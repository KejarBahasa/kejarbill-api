package handler

import (
	"time"

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

	groupID, err := h.groupService.Create(c.Context(), userID, body.Name, body.Description)
	if err != nil {
		return err
	}

	return response.Success(c, "group created", fiber.Map{
		"group_id": groupID,
	})
}

func (h *GroupHandler) GetDetailByID(c fiber.Ctx) error {
	var params dto.GroupIDParams
	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	requesterUserID := security.GetUserID(c)

	group, err := h.groupService.GetDetailByID(c.Context(), requesterUserID, params.GroupID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	result := &dto.GroupDetailResponse{
		ID:                group.ID,
		Name:              group.Name,
		Description:       group.Description,
		TotalMembers:      group.TotalMembers,
		TotalParticipants: group.TotalParticipants,
		TotalExpenses:     group.TotalExpenses,
		CreatedAt:         group.CreatedAt.Format(time.RFC3339),
	}

	return response.Success(c, "group detail fetched", result)
}
