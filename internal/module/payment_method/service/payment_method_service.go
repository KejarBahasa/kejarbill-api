package service

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/repository"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentMethodService struct {
	db *pgxpool.Pool

	encryption *security.Encryption

	paymentMethodRepo *repository.PaymentMethodRepository
}

func NewPaymentMethodService(
	db *pgxpool.Pool,
	encryption *security.Encryption,
	paymentMethodRepo *repository.PaymentMethodRepository,
) *PaymentMethodService {

	return &PaymentMethodService{
		db: db,

		encryption: encryption,

		paymentMethodRepo: paymentMethodRepo,
	}
}

func (s *PaymentMethodService) Create(ctx context.Context, userID string, req *dto.CreatePaymentMethodRequest) (string, error) {
	var encryptedAccountNumber []byte

	if req.AccountNumber != "" {
		ciphertext, err := s.encryption.Encrypt(req.AccountNumber)
		if err != nil {
			return "", err
		}

		encryptedAccountNumber = ciphertext
	}

	paymentMethod := &entity.PaymentMethod{
		UserID:        userID,
		MethodType:    req.MethodType,
		ProviderName:  req.ProviderName,
		AccountName:   utils.PtrOrNil(req.AccountName),
		AccountNumber: encryptedAccountNumber,
		QRImageURL:    utils.PtrOrNil(req.QRImageURL),
		Visibility:    req.Visibility,
		IsDefault:     req.IsDefault,
	}

	var paymentMethodID string
	err := database.WithTransaction(ctx, s.db, func(tx database.PgxExt) error {
		if req.IsDefault {
			err := s.paymentMethodRepo.ClearDefault(ctx, tx, userID)
			if err != nil {
				return err
			}
		}

		newPaymentMethodID, err := s.paymentMethodRepo.Create(ctx, tx, paymentMethod)
		if err != nil {
			return err
		}

		paymentMethodID = newPaymentMethodID

		return nil
	})

	return paymentMethodID, err
}

func (s *PaymentMethodService) FindMyPaymentMethods(ctx context.Context, userID string) ([]dto.PaymentMethodResponse, error) {
	paymentMethods, err := s.paymentMethodRepo.FindByUserID(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.PaymentMethodResponse, 0, len(paymentMethods))

	for _, paymentMethod := range paymentMethods {
		var accountNumber *string

		if len(paymentMethod.AccountNumber) > 0 {
			decrypted, err := s.encryption.Decrypt(paymentMethod.AccountNumber)
			if err != nil {
				return nil, err
			}

			accountNumber = &decrypted
		}

		result = append(result, dto.PaymentMethodResponse{
			ID:            paymentMethod.ID,
			MethodType:    paymentMethod.MethodType,
			ProviderName:  paymentMethod.ProviderName,
			AccountName:   paymentMethod.AccountName,
			AccountNumber: accountNumber,
			QRImageURL:    paymentMethod.QRImageURL,
			Visibility:    paymentMethod.Visibility,
			IsDefault:     paymentMethod.IsDefault,
			IsVerified:    paymentMethod.IsVerified,
		})
	}

	return result, nil
}
