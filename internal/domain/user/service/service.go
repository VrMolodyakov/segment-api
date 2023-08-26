package service

import (
	"context"
	"errors"

	"github.com/VrMolodyakov/segment-api/internal/domain/user/model"
	"github.com/VrMolodyakov/segment-api/pkg/logging"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExist = errors.New("user already exist")

type UserRepository interface {
	Create(ctx context.Context, user model.User) (int64, error)
	Get(ctx context.Context, userID int64) (model.User, error)
}

type service struct {
	logger logging.Logger
	user   UserRepository
}

func New(user UserRepository, logger logging.Logger) *service {
	return &service{
		user:   user,
		logger: logger,
	}
}

func (s *service) CreateUser(ctx context.Context, user model.User) (int64, error) {
	s.logger.Debugf("try to create user with email : %s", user.Email)
	return s.user.Create(ctx, user)
}

func (s *service) GetUser(ctx context.Context, userID int64) (model.User, error) {
	s.logger.Debugf("try to get user with id: %s", userID)
	return s.user.Get(ctx, userID)
}
