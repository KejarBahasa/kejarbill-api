package handler

import (
	"time"

	activityDto "github.com/KejarBahasa/kejarbill-api/internal/module/activity/dto"
	activityServicePkg "github.com/KejarBahasa/kejarbill-api/internal/module/activity/service"

	groupDto "github.com/KejarBahasa/kejarbill-api/internal/module/group/dto"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/request"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/gofiber/fiber/v3"
)

type ActivityHandler struct {
	activityService *activityServicePkg.ActivityService
}

func NewActivityHandler(
	activityService *activityServicePkg.ActivityService,
) *ActivityHandler {

	return &ActivityHandler{
		activityService: activityService,
	}
}

func (h *ActivityHandler) GetByGroupID(c fiber.Ctx) error {
	var params groupDto.GroupIDParams
	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	requesterUserID := security.GetUserID(c)

	activities, err := h.activityService.GetByGroupID(c.Context(), requesterUserID, params.GroupID)

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	result := make([]activityDto.ActivityTimelineResponse, 0, len(activities))

	for _, activity := range activities {
		item := activityDto.ActivityTimelineResponse{
			Type:      activity.Type,
			CreatedAt: activity.CreatedAt.Format(time.RFC3339),
		}

		if activity.Expense != nil {
			item.Expense = &activityDto.ActivityExpenseResponse{
				ID:               activity.Expense.ID,
				Title:            activity.Expense.Title,
				TotalAmount:      activity.Expense.TotalAmount,
				PayerDisplayName: activity.Expense.PayerDisplayName,
			}
		}

		if activity.Settlement != nil {
			item.Settlement = &activityDto.ActivitySettlementResponse{
				ID:              activity.Settlement.ID,
				Amount:          activity.Settlement.Amount,
				FromDisplayName: activity.Settlement.FromDisplayName,
				ToDisplayName:   activity.Settlement.ToDisplayName,
			}
		}

		result = append(result, item)
	}

	return response.Success(c, "activities fetched", fiber.Map{
		"activities": result,
	})
}
