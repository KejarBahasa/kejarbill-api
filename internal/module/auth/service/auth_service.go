package service

import (
	"context"
	"errors"
	"time"

	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/repository"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/utils"

	"github.com/google/uuid"
)

type AuthService struct {
	authRepo             *repository.AuthRepository
	pasetoMaker          *security.PasetoMaker
	sessionStore         *security.SessionStore
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewAuthService(
	authRepo *repository.AuthRepository,
	pasetoMaker *security.PasetoMaker,
	sessionStore *security.SessionStore,
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
) *AuthService {
	return &AuthService{
		authRepo:             authRepo,
		pasetoMaker:          pasetoMaker,
		sessionStore:         sessionStore,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) error {
	existingUser, _ := s.authRepo.FindByEmail(ctx, req.Email)
	if existingUser != nil {
		return errors.New("email already registered")
	}

	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return err
	}

	user := &entity.User{
		ID:           uuid.NewString(),
		Username:     req.Username,
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: hashedPassword,
	}

	return s.authRepo.CreateUser(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, clientInfo *utils.ClientInfo) (*dto.AuthResponse, error) {
	user, err := s.authRepo.AuthLogin(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	err = security.CheckPassword(req.Password, user.PasswordHash)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	accessToken, _, err := s.pasetoMaker.CreateToken(user.ID, security.TokenTypeAccess, user.TokenVersion, s.accessTokenDuration)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshPayload, err := s.pasetoMaker.CreateToken(user.ID, security.TokenTypeRefresh, user.TokenVersion, s.refreshTokenDuration)
	if err != nil {
		return nil, err
	}

	err = s.sessionStore.Set(
		ctx,
		&security.Session{
			TokenID:    refreshPayload.TokenID,
			UserID:     user.ID,
			UserAgent:  clientInfo.UserAgent,
			IPAddress:  clientInfo.IPAddress,
			ClientType: clientInfo.ClientType,
			ExpiredAt:  refreshPayload.ExpiredAt.Unix(),
		},
		s.refreshTokenDuration,
	)
	if err != nil {
		return nil, err
	}

	s.authRepo.UpdateLastLogin(ctx, user.ID)

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string, clientInfo *utils.ClientInfo) (*dto.AuthResponse, error) {
	payload, err := s.pasetoMaker.VerifyToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if payload.TokenType != security.TokenTypeRefresh {
		return nil, errors.New("invalid token type")
	}

	sess, err := s.sessionStore.Get(ctx, payload.TokenID)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if sess == nil {
		return nil, errors.New("session expired")
	}

	if sess.UserID != payload.UserID {
		return nil, errors.New("invalid refresh token")
	}

	if sess.ExpiredAt < time.Now().Unix() {
		return nil, errors.New("session expired")
	}

	newAccessToken, _, err := s.pasetoMaker.CreateToken(payload.UserID, security.TokenTypeAccess, payload.TokenVersion, s.accessTokenDuration)
	if err != nil {
		return nil, err
	}

	newRefreshToken, refreshPayload, err := s.pasetoMaker.CreateToken(payload.UserID, security.TokenTypeRefresh, payload.TokenVersion, s.refreshTokenDuration)
	if err != nil {
		return nil, err
	}

	// Rotate session
	err = s.sessionStore.Delete(ctx, payload.TokenID)
	if err != nil {
		return nil, err
	}

	err = s.sessionStore.Set(
		ctx,
		&security.Session{
			TokenID:    refreshPayload.TokenID,
			UserID:     payload.UserID,
			UserAgent:  clientInfo.UserAgent,
			IPAddress:  clientInfo.IPAddress,
			ClientType: clientInfo.ClientType,
			ExpiredAt:  refreshPayload.ExpiredAt.Unix(),
		},
		s.refreshTokenDuration,
	)

	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	payload, err := s.pasetoMaker.VerifyToken(refreshToken)
	if err != nil {
		return err
	}

	if payload.TokenType != security.TokenTypeRefresh {
		return errors.New("invalid token type")
	}

	return s.sessionStore.Delete(ctx, payload.TokenID)
}
