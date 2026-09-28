package handler

import (
	"errors"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	groupMemberConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/constants"
	groupParticipantConstants "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/constants"
	groupParticipantDto "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/dto"
	groupParticipantServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/service"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/request"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/gofiber/fiber/v3"
)

type GroupParticipantHandler struct {
	groupParticipantService *groupParticipantServicePkg.GroupParticipantService
}

func NewGroupParticipantHandler(
	groupParticipantService *groupParticipantServicePkg.GroupParticipantService,
) *GroupParticipantHandler {

	return &GroupParticipantHandler{
		groupParticipantService: groupParticipantService,
	}
}

func (h *GroupParticipantHandler) CreateGuestParticipants(c fiber.Ctx) error {
	var params groupParticipantDto.CreateGuestParticipantsParams
	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	var body groupParticipantDto.CreateGuestParticipantsBody
	if err := request.ValidateBody(c, &body); err != nil {
		return request.HandleValidationError(c, err)
	}

	requesterUserID := security.GetUserID(c)

	err := h.groupParticipantService.CreateGuestParticipants(c.Context(), requesterUserID, params.GroupID, &body)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return response.Success[any](c, "guest participants created", nil)
}

func (h *GroupParticipantHandler) GetByGroupID(c fiber.Ctx) error {
	var params groupParticipantDto.CreateGuestParticipantsParams
	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	requesterUserID := security.GetUserID(c)

	participants, err := h.groupParticipantService.GetByGroupID(c.Context(), requesterUserID, params.GroupID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return response.Success(c, "participants fetched", fiber.Map{
		"participants": participants,
	})
}

func (h *GroupParticipantHandler) ClaimGuest(c fiber.Ctx) error {
	var params groupParticipantDto.ClaimParticipantParams
	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	var body groupParticipantDto.ClaimParticipantBody
	if err := request.ValidateBody(c, &body); err != nil {
		return request.HandleValidationError(c, err)
	}

	requesterUserID := security.GetUserID(c)

	err := h.groupParticipantService.ClaimGuestParticipant(c.Context(), requesterUserID, params.GroupID, params.ParticipantID, body.UserID)
	if err != nil {
		return claimErrorToHTTP(err)
	}

	return response.Success[any](c, "participant claimed", nil)
}

func claimErrorToHTTP(err error) error {
	code := fiber.StatusInternalServerError
	switch {
	case errors.Is(err, expenseConstants.ErrForbiddenGroupAccess), errors.Is(err, groupMemberConstants.ErrForbiddenGroupRole):
		code = fiber.StatusForbidden
	case errors.Is(err, groupMemberConstants.ErrAlreadyMember), errors.Is(err, groupParticipantConstants.ErrParticipantNotClaimable):
		code = fiber.StatusConflict
	case errors.Is(err, expenseConstants.ErrUserNotFound), errors.Is(err, groupMemberConstants.ErrUserNotFound),
		errors.Is(err, groupParticipantConstants.ErrParticipantNotFound), errors.Is(err, expenseConstants.ErrGroupNotFound):
		code = fiber.StatusNotFound
	}

	return fiber.NewError(code, err.Error())
}
