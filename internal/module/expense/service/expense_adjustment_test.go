package service

import (
	"errors"
	"testing"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
)

func TestCalculateExpenseAdjustment(t *testing.T) {
	tests := []struct {
		name          string
		subtotal      int64
		discountType  string
		discountValue int64
		wantAmount    int64
		wantTotal     int64
	}{
		{name: "amount", subtotal: 100_000, discountType: "amount", discountValue: 10_000, wantAmount: 10_000, wantTotal: 90_000},
		{name: "percentage", subtotal: 332_000, discountType: "percentage", discountValue: 10, wantAmount: 33_200, wantTotal: 298_800},
		{name: "zero discount", subtotal: 100_000, discountValue: 0, wantAmount: 0, wantTotal: 100_000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateExpenseAdjustment(tt.subtotal, tt.discountType, tt.discountValue)
			if err != nil {
				t.Fatalf("calculateExpenseAdjustment() error = %v", err)
			}
			if got.DiscountAmount != tt.wantAmount || got.TotalAmount != tt.wantTotal {
				t.Fatalf("adjustment = %+v, want amount %d total %d", got, tt.wantAmount, tt.wantTotal)
			}
		})
	}
}

func TestAllocateDiscountProportionally(t *testing.T) {
	got := applyDiscountToShares([]int64{60_000, 40_000}, 10_000)
	if got[0] != 54_000 || got[1] != 36_000 {
		t.Fatalf("shares = %v, want [54000 36000]", got)
	}

	got = applyDiscountToShares([]int64{1, 1, 1}, 1)
	if got[0] != 0 || got[1] != 1 || got[2] != 1 {
		t.Fatalf("rounded shares = %v, want [0 1 1]", got)
	}
}

func TestCalculateExpenseAdjustmentRejectsInvalidDiscount(t *testing.T) {
	for _, test := range []struct {
		name          string
		subtotal      int64
		discountType  string
		discountValue int64
	}{
		{name: "over amount", subtotal: 100, discountType: "amount", discountValue: 101},
		{name: "over percentage", subtotal: 100, discountType: "percentage", discountValue: 101},
		{name: "free expense", subtotal: 100, discountType: "amount", discountValue: 100},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := calculateExpenseAdjustment(test.subtotal, test.discountType, test.discountValue)
			if !errors.Is(err, expenseConstants.ErrInvalidDiscount) {
				t.Fatalf("error = %v, want invalid discount", err)
			}
		})
	}
}
