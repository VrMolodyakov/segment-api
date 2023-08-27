package membership

import (
	"context"
	"errors"
	"time"

	"github.com/VrMolodyakov/segment-api/internal/domain/segment"
	"github.com/VrMolodyakov/segment-api/pkg/logging"
)

var ErrSegmentAlreadyAssigned = errors.New("segment already assigned")

type MembershipRepository interface {
	UpdateUserSegments(ctx context.Context, userID int64, addSegments []segment.Segment, deleteSegments []string) error
	DeleteSegment(ctx context.Context, name string) error
	GetUserSegments(ctx context.Context, userID int64) ([]MembershipInfo, error)
	DeleteExpired(ctx context.Context) error
}

type Cache interface {
	Set(key int64, value []MembershipInfo, expireAt time.Duration) []MembershipInfo
	Get(key int64) ([]MembershipInfo, bool)
}

type service struct {
	logger          logging.Logger
	cache           Cache
	cacheExpiration time.Duration
	membership      MembershipRepository
}

func New(participation MembershipRepository, cache Cache, expiration time.Duration, logger logging.Logger) *service {
	return &service{
		membership:      participation,
		cache:           cache,
		cacheExpiration: expiration,
		logger:          logger,
	}
}

func (s *service) UpdateUserMembership(ctx context.Context, userID int64, addSegments []segment.Segment, deleteSegments []string) error {
	s.logger.Debugf("try to update user = %d segments %v delete segments %v", addSegments, deleteSegments)
	return s.membership.UpdateUserSegments(ctx, userID, addSegments, deleteSegments)
}

func (s *service) DeleteMembership(ctx context.Context, segmentName string) error {
	s.logger.Debugf("try to delete %s segment", segmentName)
	return s.membership.DeleteSegment(ctx, segmentName)
}

func (s *service) GetUserMembership(ctx context.Context, userID int64) ([]MembershipInfo, error) {
	s.logger.Debugf("try to get user %d segments", userID)
	if info, inCache := s.cache.Get(userID); inCache {
		return info, nil
	}
	info, err := s.membership.GetUserSegments(ctx, userID)
	if err != nil {
		s.logger.Errorf("error in getting membership info, %w", err)
		return nil, err
	}
	s.cache.Set(userID, info, s.cacheExpiration)
	return info, nil
}
