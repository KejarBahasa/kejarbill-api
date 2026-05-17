package validator

import (
	"errors"

	"github.com/KejarBahasa/kejarbill-api/internal/module/expense/dto"
)

func ValidateCreateExpenseEqual(req *dto.CreateExpenseEqualRequest) error {
	if req.TotalAmount <= 0 {
		return errors.New("invalid total amount")
	}

	if len(req.ParticipantIDs) == 0 {
		return errors.New("participants required")
	}

	return nil
}
