package service

import (
	"context"
	"errors"
	"time"

	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/entity"
	"github.com/KejarBahasa/kejarbill-api/internal/module/auth/repository"
	"github.com/KejarBahasa/kejarbill-api/internal/shared/security"

	"github.com/google/uuid"
)

type AuthService struct {
	authRepo             *repository.AuthRepository
	pasetoMaker          *security.PasetoMaker
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewAuthService(
	authRepo *repository.AuthRepository,
	pasetoMaker *security.PasetoMaker,
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
) *AuthService {
	return &AuthService{
		authRepo:             authRepo,
		pasetoMaker:          pasetoMaker,
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

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.authRepo.AuthLogin(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	err = security.CheckPassword(req.Password, user.PasswordHash)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	accessToken, _, err := s.pasetoMaker.CreateToken(user.ID, s.accessTokenDuration)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := s.pasetoMaker.CreateToken(user.ID, s.refreshTokenDuration)

	if err != nil {
		return nil, err
	}

	s.authRepo.UpdateLastLogin(ctx, user.ID)

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.AuthResponse, error) {
	payload, err := s.pasetoMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	accessToken, _, err := s.pasetoMaker.CreateToken(payload.UserID, s.accessTokenDuration)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := s.pasetoMaker.CreateToken(payload.UserID, s.refreshTokenDuration)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
