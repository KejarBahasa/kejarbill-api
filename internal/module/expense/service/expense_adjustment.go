package service

import (
	"math/big"

	expenseConstants "github.com/KejarBahasa/kejarbill-api/internal/module/expense/constants"
)

type expenseAdjustment struct {
	DiscountType   string
	DiscountAmount int64
	TotalAmount    int64
}

func calculateExpenseAdjustment(subtotal int64, discountType string, discountValue int64) (expenseAdjustment, error) {
	if subtotal <= 0 || discountValue < 0 {
		return expenseAdjustment{}, expenseConstants.ErrInvalidDiscount
	}
	if discountType == "" {
		discountType = expenseConstants.DiscountTypeAmount
	}

	var discountAmount int64
	switch discountType {
	case expenseConstants.DiscountTypeAmount:
		discountAmount = discountValue
	case expenseConstants.DiscountTypePercentage:
		if discountValue > 100 {
			return expenseAdjustment{}, expenseConstants.ErrInvalidDiscount
		}
		discountAmount = new(big.Int).Quo(
			new(big.Int).Mul(big.NewInt(subtotal), big.NewInt(discountValue)),
			big.NewInt(100),
		).Int64()
	default:
		return expenseAdjustment{}, expenseConstants.ErrInvalidDiscount
	}
	if discountAmount >= subtotal {
		return expenseAdjustment{}, expenseConstants.ErrInvalidDiscount
	}

	return expenseAdjustment{
		DiscountType:   discountType,
		DiscountAmount: discountAmount,
		TotalAmount:    subtotal - discountAmount,
	}, nil
}

// allocateDiscount distributes the discount according to each participant's
// pre-discount share. Remaining rupiah follows the input order.
func allocateDiscount(shares []int64, discountAmount int64) []int64 {
	allocations := make([]int64, len(shares))
	if discountAmount == 0 {
		return allocations
	}

	var subtotal int64
	for _, share := range shares {
		subtotal += share
	}
	if subtotal == 0 {
		return allocations
	}

	var allocated int64
	remainingNumerators := make([]*big.Int, len(shares))
	for index, share := range shares {
		numerator := new(big.Int).Mul(big.NewInt(discountAmount), big.NewInt(share))
		quotient, remainder := new(big.Int), new(big.Int)
		quotient.QuoRem(numerator, big.NewInt(subtotal), remainder)
		allocations[index] = quotient.Int64()
		allocated += allocations[index]
		remainingNumerators[index] = remainder
	}

	for index := 0; allocated < discountAmount; index++ {
		if remainingNumerators[index].Sign() > 0 {
			allocations[index]++
			allocated++
		}
		if index == len(allocations)-1 {
			index = -1
		}
	}

	return allocations
}

func applyDiscountToShares(shares []int64, discountAmount int64) []int64 {
	allocations := allocateDiscount(shares, discountAmount)
	finalShares := make([]int64, len(shares))
	for index, share := range shares {
		finalShares[index] = share - allocations[index]
	}
	return finalShares
}
