package security

import (
	"crypto/sha256"

	"golang.org/x/crypto/bcrypt"
)

func preHashPassword(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

func HashPassword(plainPassword string) (string, error) {
	hashedSHA256 := preHashPassword(plainPassword)

	hashedPassword, err := bcrypt.GenerateFromPassword(
		hashedSHA256,
		bcrypt.DefaultCost,
	)

	return string(hashedPassword), err
}

func CheckPassword(plainPassword string, hashedPassword string) error {
	hashedSHA256 := preHashPassword(plainPassword)

	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		hashedSHA256,
	)
}
