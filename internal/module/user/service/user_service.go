package service

import (
	"context"
	"time"

	"github.com/KejarBahasa/kejarbill-api/internal/module/user/dto"
	"github.com/KejarBahasa/kejarbill-api/internal/module/user/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(
	userRepo *repository.UserRepository,
) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) Me(ctx context.Context, userID string) (*dto.MeResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.MeResponse{
		ID:        user.ID,
		Name:      user.Name,
		Username:  user.Username,
		Email:     user.Email,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}
