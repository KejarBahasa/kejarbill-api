package repository

import (
	"context"
	"errors"

	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/constants"
	"github.com/KejarBahasa/kejarbill-api/internal/module/settlement/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/database"
	"github.com/jackc/pgx/v5"
)

type SettlementRepository struct{}

func NewSettlementRepository() *SettlementRepository {
	return &SettlementRepository{}
}

func (r *SettlementRepository) Create(ctx context.Context, db database.PgxExt, settlement *entity.Settlement) (string, error) {
	query := `
		INSERT INTO settlements (
			group_id,
			from_participant_id,
			to_participant_id,
			amount,
			payment_channel,
			payment_method_id,
			status,
			notes,
			paid_at,
			created_by
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10
		)
		RETURNING id
	`

	var settlementID string
	err := db.QueryRow(
		ctx,
		query,

		settlement.GroupID,
		settlement.FromParticipantID,
		settlement.ToParticipantID,
		settlement.Amount,
		settlement.PaymentChannel,
		settlement.PaymentMethodID,
		settlement.Status,
		settlement.Notes,
		settlement.PaidAt,
		settlement.CreatedBy,
	).Scan(&settlementID)

	return settlementID, err
}

func (r *SettlementRepository) CreateIdempotencyKey(ctx context.Context, db database.PgxExt, idempotencyKey *entity.IdempotencyKey) (bool, error) {
	query := `
		INSERT INTO idempotency_keys (
			user_id,
			idempotency_key,
			request_hash,
			expired_at
		)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (idempotency_key) DO NOTHING
	`

	commandTag, err := db.Exec(
		ctx,
		query,
		idempotencyKey.UserID,
		idempotencyKey.Key,
		idempotencyKey.RequestHash,
		idempotencyKey.ExpiredAt,
	)
	if err != nil {
		return false, err
	}

	return commandTag.RowsAffected() == 1, nil
}

func (r *SettlementRepository) FindIdempotencyKeyByKey(ctx context.Context, db database.PgxExt, key string) (*entity.IdempotencyKey, error) {
	query := `
		SELECT
			user_id,
			idempotency_key,
			request_hash,
			response_code,
			response_body::TEXT,
			expired_at
		FROM idempotency_keys
		WHERE idempotency_key = $1
	`

	var idempotencyKey entity.IdempotencyKey
	err := db.QueryRow(ctx, query, key).Scan(
		&idempotencyKey.UserID,
		&idempotencyKey.Key,
		&idempotencyKey.RequestHash,
		&idempotencyKey.ResponseCode,
		&idempotencyKey.ResponseBody,
		&idempotencyKey.ExpiredAt,
	)
	if err != nil {
		return nil, err
	}

	return &idempotencyKey, nil
}

func (r *SettlementRepository) SaveIdempotencyResponse(ctx context.Context, db database.PgxExt, key string, responseCode int, responseBody string) error {
	query := `
		UPDATE idempotency_keys
		SET
			response_code = $2,
			response_body = $3::JSONB
		WHERE idempotency_key = $1
	`

	_, err := db.Exec(ctx, query, key, responseCode, responseBody)

	return err
}

func (r *SettlementRepository) CountByGroupID(ctx context.Context, db database.PgxExt, groupID string) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM settlements
		WHERE group_id = $1
	`

	var total int64
	err := db.QueryRow(ctx, query, groupID).Scan(&total)

	return total, err
}

func (r *SettlementRepository) FindByGroupID(ctx context.Context, db database.PgxExt, groupID string, limit, offset int) ([]entity.SettlementTimeline, error) {
	query := `
		SELECT
			s.id,
			s.amount,
			s.created_at,

			fp.id,
			fp.display_name,

			tp.id,
			tp.display_name
		FROM settlements s
		INNER JOIN group_participants fp
			ON fp.id = s.from_participant_id
		INNER JOIN group_participants tp
			ON tp.id = s.to_participant_id
		WHERE s.group_id = $1
		ORDER BY s.paid_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(ctx, query, groupID, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	settlements := make([]entity.SettlementTimeline, 0)

	for rows.Next() {
		var settlement entity.SettlementTimeline
		err := rows.Scan(
			&settlement.ID,
			&settlement.Amount,
			&settlement.SettlementDate,
			&settlement.FromParticipantID,
			&settlement.FromDisplayName,
			&settlement.ToParticipantID,
			&settlement.ToDisplayName,
		)
		if err != nil {
			return nil, err
		}

		settlements = append(settlements, settlement)
	}

	return settlements, nil
}

func (r *SettlementRepository) FindDetailByID(ctx context.Context, db database.PgxExt, settlementID string) (*entity.SettlementDetail, error) {
	query := `
		SELECT
			s.id,
			s.group_id,
			s.amount,
			s.payment_channel,
			s.payment_method_id,
			s.notes,
			s.paid_at,
			s.status,

			fp.id,
			fp.display_name,

			tp.id,
			tp.display_name,

			pm.provider_name,
			pm.method_type

		FROM settlements s

		INNER JOIN group_participants fp
			ON fp.id = s.from_participant_id

		INNER JOIN group_participants tp
			ON tp.id = s.to_participant_id

		LEFT JOIN payment_methods pm
			ON pm.id = s.payment_method_id

		WHERE s.id = $1
	`

	var detail entity.SettlementDetail
	err := db.QueryRow(ctx, query, settlementID).Scan(
		&detail.ID,
		&detail.GroupID,
		&detail.Amount,
		&detail.PaymentChannel,
		&detail.PaymentMethodID,
		&detail.Notes,
		&detail.PaidAt,
		&detail.Status,
		&detail.FromParticipantID,
		&detail.FromDisplayName,
		&detail.ToParticipantID,
		&detail.ToDisplayName,
		&detail.ProviderName,
		&detail.MethodType,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, constants.ErrSettlementNotFound
		}
		return nil, err
	}

	return &detail, nil
}
