package user

import (
	"context"
	"errors"

	"github.com/VrMolodyakov/segment-api/pkg/logging"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExist = errors.New("user already exist")

type UserRepository interface {
	Get(ctx context.Context, userID int64) (User, error)
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

func (s *service) GetUser(ctx context.Context, userID int64) (User, error) {
	s.logger.Debugf("try to get user with id: %s", userID)
	return s.user.Get(ctx, userID)
}
