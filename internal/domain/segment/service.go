package segment

import (
	"context"
	"errors"

	"github.com/VrMolodyakov/segment-api/pkg/logging"
)

var ErrSegmentNotFound = errors.New("sergment not found")
var ErrSegmentAlreadyExists = errors.New("segment already exists")

type SegmentRepository interface {
	Create(ctx context.Context, name string, percentage int) (int64, error)
	Get(ctx context.Context, name string) (SegmentInfo, error)
	GetAll(ctx context.Context) ([]SegmentInfo, error)
}

type service struct {
	logger  logging.Logger
	segment SegmentRepository
}

func New(segment SegmentRepository, logger logging.Logger) *service {
	return &service{
		segment: segment,
		logger:  logger,
	}
}

func (s *service) CreateSegment(ctx context.Context, name string, percentage int) (int64, error) {
	s.logger.Debugf("try to create segment , name : %s", name)
	if _, err := s.segment.Get(ctx, name); err != ErrSegmentNotFound {
		if err == nil {
			s.logger.Error("segment %s already exists", name)
			return 0, ErrSegmentAlreadyExists
		}
		return 0, err

	}
	id, err := s.segment.Create(ctx, name, percentage)
	if err != nil {
		s.logger.Error("cannot create segment %s", err.Error())
	}
	return id, err
}

func (s *service) GetAllSegments(ctx context.Context) ([]SegmentInfo, error) {
	return s.segment.GetAll(ctx)
}
