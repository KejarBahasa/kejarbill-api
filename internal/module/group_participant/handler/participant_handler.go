package handler

import (
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

	return response.Success(c, "guest participants created", nil)
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

	result := make([]groupParticipantDto.ParticipantResponse, 0, len(participants))
	for _, participant := range participants {
		result = append(result, groupParticipantDto.ParticipantResponse{
			ID:              participant.ID,
			UserID:          participant.UserID,
			ParticipantType: participant.ParticipantType,
			DisplayName:     participant.DisplayName,
		})
	}

	return response.Success(c, "participants fetched", fiber.Map{
		"participants": result,
	})
}
