package security

import (
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	paseto "aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
)

type PasetoMaker struct {
	parser paseto.Parser
	key    paseto.V4SymmetricKey
}

func NewPasetoMaker(secretStr string) (*PasetoMaker, error) {
	rawSecret, err := base64.StdEncoding.DecodeString(secretStr)
	if err != nil {
		return nil, errors.New("invalid secret key encoding")
	}

	key, err := paseto.V4SymmetricKeyFromBytes(rawSecret)
	if err != nil {
		return nil, err
	}

	parser := paseto.NewParser()

	return &PasetoMaker{
		parser: parser,
		key:    key,
	}, nil
}

func (p *PasetoMaker) CreateToken(userID, tokenType string, tokenVersion int16, duration time.Duration) (string, *Payload, error) {
	now := time.Now()

	payload := &Payload{
		TokenID:      uuid.NewString(),
		UserID:       userID,
		TokenType:    tokenType,
		TokenVersion: tokenVersion,
		ExpiredAt:    now.Add(duration),
	}

	token := paseto.NewToken()

	token.SetIssuedAt(now)
	token.SetExpiration(payload.ExpiredAt)

	token.SetString(ContextTokenID, payload.TokenID)
	token.SetString(ContextTokenType, payload.TokenType)
	token.SetString(ContextUserID, payload.UserID)
	token.SetString(ContextTokenVersion, strconv.Itoa(int(payload.TokenVersion)))

	tokenStr := token.V4Encrypt(p.key, nil)

	return tokenStr, payload, nil
}

func (p *PasetoMaker) VerifyToken(tokenStr string) (*Payload, error) {
	token, err := p.parser.ParseV4Local(
		p.key,
		tokenStr,
		nil,
	)

	if err != nil {
		return nil, err
	}

	exp, err := token.GetExpiration()
	if err != nil {
		return nil, err
	}

	if time.Now().After(exp) {
		return nil, errors.New("token expired")
	}

	tokenId, err := token.GetString(ContextTokenID)
	if err != nil {
		return nil, err
	}

	tokenType, err := token.GetString(ContextTokenType)
	if err != nil {
		return nil, err
	}

	userID, err := token.GetString(ContextUserID)
	if err != nil {
		return nil, err
	}

	tokenVersionStr, err := token.GetString(ContextTokenVersion)
	if err != nil {
		return nil, err
	}

	tokenVersion, err := strconv.Atoi(tokenVersionStr)
	if err != nil {
		return nil, err
	}

	return &Payload{
		TokenID:      tokenId,
		UserID:       userID,
		TokenType:    tokenType,
		TokenVersion: int16(tokenVersion),
		ExpiredAt:    exp,
	}, nil
}
