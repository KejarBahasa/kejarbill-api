package service

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/entity"
	paymentMethodConst "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
)

func (s *SettlementService) validatePaymentMethod(ctx context.Context, db database.PgxExt, request dto.CreateSettlementBody, toParticipant *entity.GroupParticipant) error {
	switch request.PaymentChannel {

	case constants.PaymentChannelCash:
		if request.PaymentMethodID != "" {
			return constants.ErrPaymentMethodMustBeEmpty
		}

	case constants.PaymentChannelBankTransfer,
		constants.PaymentChannelEwallet:

		if request.PaymentMethodID == "" {
			return constants.ErrPaymentMethodRequired
		}

		pm, err := s.paymentMethodRepo.FindByID(ctx, db, request.PaymentMethodID)
		if err != nil {
			return err
		}

		if pm.Status != paymentMethodConst.StatusActive {
			return paymentMethodConst.ErrPaymentMethodInactive
		}

		if toParticipant.UserID == nil || pm.UserID != *toParticipant.UserID {
			return constants.ErrPaymentMethodNotOwned
		}

		expected := map[string]string{
			constants.PaymentChannelBankTransfer: paymentMethodConst.MethodTypeBank,
			constants.PaymentChannelEwallet:      paymentMethodConst.MethodTypeEWallet,
		}

		if pm.MethodType != expected[request.PaymentChannel] {
			return constants.ErrInvalidPaymentMethodType
		}
	}

	return nil
}
