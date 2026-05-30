package utils

func MaskAccountNumber(accountNumber string) string {
	if len(accountNumber) <= 4 {
		return accountNumber
	}

	return "****" + accountNumber[len(accountNumber)-4:]
}
