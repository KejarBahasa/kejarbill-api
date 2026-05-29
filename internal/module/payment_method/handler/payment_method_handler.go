package handler

import (
	"github.com/gofiber/fiber/v3"

	"github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/service"

	"github.com/KejarBahasa/kejarbill-api/internal/shared/request"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/response"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"
)

type PaymentMethodHandler struct {
	paymentMethodService *service.PaymentMethodService
}

func NewPaymentMethodHandler(
	paymentMethodService *service.PaymentMethodService,
) *PaymentMethodHandler {
	return &PaymentMethodHandler{
		paymentMethodService: paymentMethodService,
	}
}

func (h *PaymentMethodHandler) Create(c fiber.Ctx) error {
	var req dto.CreatePaymentMethodRequest
	if err := request.ValidateBody(c, &req); err != nil {
		return request.HandleValidationError(c, err)
	}

	userID := security.GetUserID(c)

	paymentMethodID, err := h.paymentMethodService.Create(c.Context(), userID, &req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success[any](c, "payment method created", fiber.Map{
		"payment_method_id": paymentMethodID,
	}, fiber.StatusCreated)
}

func (h *PaymentMethodHandler) FindMine(c fiber.Ctx) error {
	userID := security.GetUserID(c)

	paymentMethods, err := h.paymentMethodService.FindMyPaymentMethods(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, "payment methods fetched", fiber.Map{
		"payment_methods": paymentMethods,
	})
}

func (h *PaymentMethodHandler) SetDefault(c fiber.Ctx) error {
	var params dto.PaymentMethodIDParams

	if err := request.ValidatePathParams(c, &params); err != nil {
		return request.HandleValidationError(c, err)
	}

	userID := security.GetUserID(c)

	err := h.paymentMethodService.SetDefault(c.Context(), userID, params.PaymentMethodID)
	if err != nil {
		return err
	}

	return response.Success[any](c, "payment method updated", nil)
}
