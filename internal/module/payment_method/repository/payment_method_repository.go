package repository

import (
	"context"

	"github.com/KejarBahasa/kejarbill-api/internal/module/payment_method/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
)

type PaymentMethodRepository struct {
}

func NewPaymentMethodRepository() *PaymentMethodRepository {
	return &PaymentMethodRepository{}
}

func (r *PaymentMethodRepository) Create(ctx context.Context, db database.PgxExt, paymentMethod *entity.PaymentMethod) (string, error) {
	query := `
		INSERT INTO payment_methods (
			user_id,
			method_type,
			provider_name,
			account_name,
			account_number,
			qr_image_url,
			visibility,
			is_default
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8
		)
		RETURNING id
	`

	var id string
	err := db.QueryRow(
		ctx,
		query,
		paymentMethod.UserID,
		paymentMethod.MethodType,
		paymentMethod.ProviderName,
		paymentMethod.AccountName,
		paymentMethod.AccountNumber,
		paymentMethod.QRImageURL,
		paymentMethod.Visibility,
		paymentMethod.IsDefault,
	).Scan(&id)

	return id, err
}

func (r *PaymentMethodRepository) ClearDefault(ctx context.Context, db database.PgxExt, userID string) error {
	query := `
		UPDATE payment_methods
		SET is_default = FALSE
		WHERE user_id = $1
			AND is_default IS TRUE
			AND deleted_at IS NULL
	`

	_, err := db.Exec(ctx, query, userID)

	return err
}

func (r *PaymentMethodRepository) FindByUserID(ctx context.Context, db database.PgxExt, userID string) ([]entity.PaymentMethod, error) {
	query := `
		SELECT
			id,
			user_id,
			method_type,
			provider_name,
			account_name,
			account_number,
			qr_image_url,
			visibility,
			is_default,
			is_verified,
			status
		FROM payment_methods
		WHERE user_id = $1
			AND deleted_at IS NULL
			AND status <> 'deleted'
		ORDER BY is_default DESC, created_at ASC
	`

	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]entity.PaymentMethod, 0)

	for rows.Next() {
		var item entity.PaymentMethod

		err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.MethodType,
			&item.ProviderName,
			&item.AccountName,
			&item.AccountNumber,
			&item.QRImageURL,
			&item.Visibility,
			&item.IsDefault,
			&item.IsVerified,
			&item.Status,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	return result, rows.Err()
}

func (r *PaymentMethodRepository) ExistsByIDAndUserID(ctx context.Context, db database.PgxExt, userID string, paymentMethodID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM payment_methods
			WHERE id = $1
				AND user_id = $2
				AND deleted_at IS NULL
				AND status != 'deleted'
		)
	`

	var exists bool
	err := db.QueryRow(ctx, query, paymentMethodID, userID).Scan(&exists)

	return exists, err
}

func (r *PaymentMethodRepository) SetDefault(ctx context.Context, db database.PgxExt, paymentMethodID string) error {
	query := `
		UPDATE payment_methods
		SET is_default = TRUE
		WHERE id = $1
	`

	_, err := db.Exec(ctx, query, paymentMethodID)

	return err
}
