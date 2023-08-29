package cleaner

import (
	"context"
	"time"

	"github.com/VrMolodyakov/segment-api/pkg/logging"
)

type MembershipRepository interface {
	DeleteExpired(ctx context.Context) error
}

type service struct {
	logger     logging.Logger
	membership MembershipRepository
}

func New(membership MembershipRepository, logger logging.Logger) *service {
	return &service{
		membership: membership,
		logger:     logger,
	}
}

func (s *service) Start(ctx context.Context, interval time.Duration) {
	if interval > 0 {
		go s.deleteExpired(ctx, interval)
	}
}

func (s *service) deleteExpired(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for range t.C {
		select {
		case <-ctx.Done():
			s.logger.Infof("context was closed, %s", ctx.Err().Error())
			return
		default:
			childCtx, cancel := context.WithTimeout(ctx, interval)
			if err := s.membership.DeleteExpired(childCtx); err != nil {
				s.logger.Errorf("couldn't delete expired rows, %s", err.Error())
			}
			cancel()
		}
	}
}
