package membership

import (
	"context"
	"errors"
	"time"

	"github.com/VrMolodyakov/segment-api/internal/domain/segment"
	"github.com/VrMolodyakov/segment-api/internal/domain/user"

	"github.com/VrMolodyakov/segment-api/pkg/logging"
	"github.com/VrMolodyakov/segment-api/pkg/random"
)

var (
	ErrSegmentAlreadyAssigned = errors.New("segment already assigned")
	ErrSegmentNotAssigned     = errors.New("segment not assigned")
	ErrSegmentNotExists       = errors.New("not all segments were found")
	ErrEmptyData              = errors.New("data for updating and for deletion were not provided")
	ErrIncorrectData          = errors.New("attempt to add and remove the same segment")
)

type MembershipRepository interface {
	UpdateUserSegments(ctx context.Context, userID int64, addSegments []segment.Segment, deleteSegments []string) error
	DeleteSegment(ctx context.Context, name string) error
	GetUserSegments(ctx context.Context, userID int64) ([]MembershipInfo, error)
	CreateUser(ctx context.Context, user user.User, hitPercentage int) (int64, error)
}

type Cache interface {
	Set(key int64, value []MembershipInfo, expireAt time.Duration) []MembershipInfo
	Get(key int64) ([]MembershipInfo, bool)
}

type service struct {
	logger          logging.Logger
	cache           Cache
	cacheExpiration time.Duration
	generator       random.Rand
	membership      MembershipRepository
}

func New(
	membership MembershipRepository,
	cache Cache,
	expiration time.Duration,
	generator random.Rand,
	logger logging.Logger,
) *service {

	return &service{
		membership:      membership,
		cache:           cache,
		generator:       generator,
		cacheExpiration: expiration,
		logger:          logger,
	}
}

func (s *service) DeleteMembership(ctx context.Context, segmentName string) error {
	s.logger.Debugf("try to delete %s segment", segmentName)
	err := s.membership.DeleteSegment(ctx, segmentName)
	if err != nil {
		s.logger.Errorf("cannot delete %s segment due to %s", segmentName, err.Error())
	}
	return err
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

func (s *service) CreateUser(ctx context.Context, user user.User) (int64, error) {
	s.logger.Debugf("try to create user %s ", user.Email)
	if err := user.Valid(); err != nil {
		s.logger.Errorf("invalid email %s", user.Email)
		return 0, err
	}
	id, err := s.membership.CreateUser(ctx, user, s.generator.Next())
	if err != nil {
		s.logger.Errorf("error in creating user, %s", err.Error())
	}
	return id, err
}

func (s *service) UpdateUserMembership(
	ctx context.Context,
	userID int64,
	addSegments []segment.Segment,
	deleteSegments []string,
) error {
	s.logger.Debugf("try to update user = %d segments %v delete segments %v", addSegments, deleteSegments)
	if err := validateUpdatedData(addSegments, deleteSegments); err != nil {
		return err
	}
	err := s.membership.UpdateUserSegments(ctx, userID, addSegments, deleteSegments)
	if err != nil {
		s.logger.Errorf("error in updating user segments, %s", err.Error())
	}
	return err
}

func validateUpdatedData(add []segment.Segment, delete []string) error {
	if len(add) == 0 && len(delete) == 0 {
		return ErrEmptyData
	}
	set := make(map[string]struct{}, max(len(add), len(delete)))
	for i := range add {
		set[add[i].Name] = struct{}{}
	}
	for i := range delete {
		if _, ok := set[delete[i]]; ok {
			return ErrIncorrectData
		}
	}
	return nil
}

func max(a, b int) int {
	if a < b {
		return a
	}
	return b
}
