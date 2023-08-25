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
	Create(ctx context.Context, user model.User) (int, error)
	Get(ctx context.Context, userID int) (model.User, error)
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

func (s *service) CreateUser(ctx context.Context, user model.User) {

}
