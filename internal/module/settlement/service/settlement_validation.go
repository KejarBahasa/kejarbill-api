package service

import (
	"context"

	groupMemberConst "github.com/KejarBahasa/kejarbill-api/internal/module/group_member/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/group_participant/entity"
	paymentMethodConst "github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
)

func (s *SettlementService) validateSettlementSender(ctx context.Context, db database.PgxExt, requesterUserID string, groupID string, fromParticipantID string) error {
	fromParticipant, err := s.groupParticipantRepo.FindByIDAndGroupID(ctx, db, fromParticipantID, groupID)
	if err != nil {
		return err
	}
	if fromParticipant == nil {
		return constants.ErrInvalidSettlementParticipants
	}

	if fromParticipant.UserID != nil {
		if *fromParticipant.UserID != requesterUserID {
			return constants.ErrSettlementSenderNotAllowed
		}
		return nil
	}

	role, err := s.groupMemberRepo.FindActiveMemberRole(ctx, db, groupID, requesterUserID)
	if err != nil {
		return err
	}
	if !groupMemberConst.CanManage(role) {
		return constants.ErrSettlementSenderNotAllowed
	}

	return nil
}

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
