package security

import (
	"context"
	"encoding/json"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type SessionStore struct {
	rdb *goredis.Client
}

type Session struct {
	TokenID    string `json:"token_id"`
	UserID     string `json:"user_id"`
	UserAgent  string `json:"user_agent"`
	IPAddress  string `json:"ip_address"`
	ClientType string `json:"client_type"`
	ExpiredAt  int64  `json:"expired_at"`
}

func NewSessionStore(
	rdb *goredis.Client,
) *SessionStore {
	return &SessionStore{
		rdb: rdb,
	}
}

func (s *SessionStore) Set(ctx context.Context, session *Session, duration time.Duration) error {
	b, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return s.rdb.Set(ctx, "session:"+session.TokenID, b, duration).Err()
}

func (s *SessionStore) Get(ctx context.Context, tokenID string) (*Session, error) {
	result, err := s.rdb.Get(ctx, "session:"+tokenID).Result()
	if err != nil {
		return nil, err
	}

	var session Session
	err = json.Unmarshal([]byte(result), &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *SessionStore) Delete(ctx context.Context, tokenID string) error {
	return s.rdb.Del(ctx, "session:"+tokenID).Err()
}
