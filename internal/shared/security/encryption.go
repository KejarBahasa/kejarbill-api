package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type Encryption struct {
	key []byte
}

func NewEncryption(key string) (*Encryption, error) {
	decodedKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	if len(decodedKey) != 32 {
		return nil, fmt.Errorf("invalid key length")
	}

	return &Encryption{
		key: decodedKey,
	}, nil
}

func (e *Encryption) Encrypt(plaintext string) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())

	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	return []byte(encoded), nil
}

func (e *Encryption) Decrypt(ciphertext []byte) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(ciphertext))
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()

	if len(decoded) < nonceSize {
		return "", fmt.Errorf("invalid ciphertext")
	}

	nonce := decoded[:nonceSize]

	encrypted := decoded[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
