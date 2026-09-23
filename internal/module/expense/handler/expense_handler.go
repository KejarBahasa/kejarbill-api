package handler

import (
	"errors"
	"time"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/service"

	groupDto "github.com/KejarBahasa/kejarbill-api/internal/module/group/dto"

	sharedDto "github.com/KejarBahasa/kejarbill-api/internal/shared/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/request"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/gofiber/fiber/v3"
)

type ExpenseHandler struct {
	expenseService *service.ExpenseService
}

func NewExpenseHandler(
	expenseService *service.ExpenseService,
) *ExpenseHandler {

	return &ExpenseHandler{
		expenseService: expenseService,
	}
}

func (h *ExpenseHandler) CreateEqualExpense(c fiber.Ctx) error {
	var req dto.CreateExpenseEqualRequest
	if validationErr := request.ValidateBody(c, &req); validationErr != nil {
		return request.HandleValidationError(c, validationErr)
	}

	userID := security.GetUserID(c)

	expenseID, err := h.expenseService.CreateEqualExpense(c.Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, expenseConstants.ErrParticipantsRequired):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrGroupNotFound):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrForbiddenGroupAccess):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrPayerNotIncludedInParticipants):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrPayerParticipantNotInGroup):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrDuplicateParticipants):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrParticipantNotInGroup):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrInvalidExpenseDate):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		default:
			return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
		}
	}

	return response.Success(c, "expense created", fiber.Map{"id": expenseID})
}

func (h *ExpenseHandler) CreateCustomExpense(c fiber.Ctx) error {
	var req dto.CreateExpenseCustomRequest
	if err := request.ValidateBody(c, &req); err != nil {
		return request.HandleValidationError(c, err)
	}

	userID := security.GetUserID(c)

	expenseID, err := h.expenseService.CreateCustomExpense(c.Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, expenseConstants.ErrParticipantsRequired):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrGroupNotFound):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrForbiddenGroupAccess):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrPayerNotIncludedInParticipants):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrPayerParticipantNotInGroup):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrDuplicateParticipants):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrParticipantNotInGroup):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		case errors.Is(err, expenseConstants.ErrInvalidExpenseDate):
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)

		default:
			return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
		}
	}

	return response.Success(
		c,
		"custom expense created",
		fiber.Map{
			"expense_id": expenseID,
		},
		fiber.StatusCreated,
	)
}

func (h *ExpenseHandler) CreateItemizedExpense(c fiber.Ctx) error {
	var req dto.CreateExpenseItemizedRequest
	if err := request.ValidateBody(c, &req); err != nil {
		return request.HandleValidationError(c, err)
	}

	userID := security.GetUserID(c)

	expenseID, err := h.expenseService.CreateItemizedExpense(c.Context(), userID, &req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return response.Success(
		c,
		"itemized expense created",
		fiber.Map{
			"expense_id": expenseID,
		},
		fiber.StatusCreated,
	)
}

func (h *ExpenseHandler) GetByGroupID(c fiber.Ctx) error {
	var params groupDto.GroupIDParams
	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	var query sharedDto.PaginationQuery
	if err := request.ValidateQuery(c, &query); err != nil {
		return request.HandleValidationError(c, err)
	}

	requesterUserID := security.GetUserID(c)

	expenses, err := h.expenseService.GetByGroupID(c.Context(), requesterUserID, params.GroupID, query.Page, query.Limit)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	result := make([]dto.ExpenseTimelineResponse, 0, len(expenses.Expenses))
	for _, expense := range expenses.Expenses {
		result = append(result, dto.ExpenseTimelineResponse{
			ID:          expense.ID,
			Title:       expense.Title,
			Description: expense.Description,
			Currency:    expense.Currency,
			TotalAmount: expense.TotalAmount,
			ExpenseDate: expense.ExpenseDate.Format(
				time.RFC3339,
			),
			Payer: dto.ExpenseTimelinePayerResponse{
				ParticipantID: expense.PayerParticipantID,
				DisplayName:   expense.PayerDisplayName,
			},
		})
	}

	return response.SuccessWithMeta(c, "expenses fetched",
		fiber.Map{
			"expenses": result,
		},
		&response.Meta{
			Pagination: &response.PaginationMeta{
				Page:       expenses.Page,
				Limit:      expenses.Limit,
				TotalItems: expenses.TotalItems,
				TotalPages: expenses.TotalPages,
			},
		},
	)
}

func (h *ExpenseHandler) GetDetailByID(c fiber.Ctx) error {
	var params dto.GetExpenseDetailParams
	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	requesterUserID := security.GetUserID(c)

	expense, err := h.expenseService.GetDetailByID(c.Context(), requesterUserID, params.ExpenseID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	participants := make([]dto.ExpenseDetailParticipantResponse, 0, len(expense.Participants))
	items := make([]dto.ExpenseDetailItemResponse, 0, len(expense.Items))

	for _, participant := range expense.Participants {
		participants = append(participants, dto.ExpenseDetailParticipantResponse{
			ParticipantID:   participant.ParticipantID,
			DisplayName:     participant.DisplayName,
			ParticipantType: participant.ParticipantType,
			ShareAmount:     participant.ShareAmount,
		})
	}

	for _, item := range expense.Items {
		items = append(items, dto.ExpenseDetailItemResponse{
			ID:        item.ID,
			Name:      item.Name,
			Qty:       item.Qty,
			UnitPrice: item.UnitPrice,
			Subtotal:  item.Subtotal,
			Notes:     item.Notes,
		})
	}

	result := dto.ExpenseDetailResponse{
		ID:          expense.ID,
		Title:       expense.Title,
		Description: expense.Description,
		Currency:    expense.Currency,
		TotalAmount: expense.TotalAmount,
		ExpenseDate: expense.ExpenseDate.Format(
			time.RFC3339,
		),
		Payer: dto.ExpenseDetailPayerResponse{
			ParticipantID: expense.PayerParticipantID,
			DisplayName:   expense.PayerDisplayName,
		},
		Participants: participants,
		Items:        items,
	}

	return response.Success(c, "expense detail fetched", result)
}

func (h *ExpenseHandler) DeleteByID(c fiber.Ctx) error {
	var params dto.DeleteExpenseParams
	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	requesterUserID := security.GetUserID(c)

	err := h.expenseService.DeleteByID(c.Context(), requesterUserID, params.ExpenseID)

	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return response.Success[any](c, "expense deleted", nil)
}
