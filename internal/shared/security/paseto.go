package security

import (
	"encoding/base64"
	"errors"
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

func (p *PasetoMaker) CreateToken(userID string, duration time.Duration) (string, *Payload, error) {
	now := time.Now()

	payload := &Payload{
		TokenID:   uuid.NewString(),
		UserID:    userID,
		ExpiredAt: now.Add(duration),
	}

	token := paseto.NewToken()

	token.SetIssuedAt(now)
	token.SetExpiration(payload.ExpiredAt)

	token.SetString("token_id", payload.TokenID)
	token.SetString(ContextUserID, payload.UserID)

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

	tokenId, err := token.GetString("token_id")
	if err != nil {
		return nil, err
	}

	userId, err := token.GetString("user_id")
	if err != nil {
		return nil, err
	}

	return &Payload{
		TokenID:   tokenId,
		UserID:    userId,
		ExpiredAt: exp,
	}, nil
}
