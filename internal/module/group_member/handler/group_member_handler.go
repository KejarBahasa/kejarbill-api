package handler

import (
	groupMemberDto "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/dto"

	groupMemberServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/service"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/request"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/gofiber/fiber/v3"
)

type GroupMemberHandler struct {
	groupMemberService *groupMemberServicePkg.GroupMemberService
}

func NewGroupMemberHandler(
	groupMemberService *groupMemberServicePkg.GroupMemberService,
) *GroupMemberHandler {

	return &GroupMemberHandler{
		groupMemberService: groupMemberService,
	}
}

func (h *GroupMemberHandler) AddMember(c fiber.Ctx) error {
	var params groupMemberDto.AddGroupMemberParams
	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	var body groupMemberDto.AddGroupMemberBody
	if err := request.ValidateBody(c, &body); err != nil {
		return request.HandleValidationError(c, err)
	}

	requesterUserID := security.GetUserID(c)
	err := h.groupMemberService.AddMember(c.Context(), requesterUserID, params.GroupID, &body)

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return response.Success[any](c, "member added", nil)
}

func (h *GroupMemberHandler) AddMemberBulk(c fiber.Ctx) error {
	var params groupMemberDto.AddGroupMemberParams
	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	var body groupMemberDto.AddGroupMembersBulkBody
	if err := request.ValidateBody(c, &body); err != nil {
		return request.HandleValidationError(c, err)
	}

	userID := security.GetUserID(c)

	err := h.groupMemberService.AddMembersBulk(c.Context(), userID, params.GroupID, &body)

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return response.Success[any](c, "members added", nil)
}
