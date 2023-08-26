package service

import (
	"context"
	"errors"

	"github.com/VrMolodyakov/segment-api/internal/domain/segment/model"
	"github.com/VrMolodyakov/segment-api/pkg/logging"
)

var ErrSegmentNotFound = errors.New("sergment not found")
var ErrSegmentAlreadyExists = errors.New("segment already exists")

type SegmentRepository interface {
	Create(ctx context.Context, name string) (int64, error)
	Get(ctx context.Context, name string) (model.SegmentInfo, error)
	GetAll(ctx context.Context) ([]model.SegmentInfo, error)
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

func (s *service) CreateSegment(ctx context.Context, name string) (int64, error) {
	s.logger.Debugf("try to create segment , name : %s", name)
	if _, err := s.segment.Get(ctx, name); err != ErrSegmentNotFound {
		if err == nil {
			s.logger.Error("segment %s already exists", name)
			return 0, ErrSegmentAlreadyExists
		}
		return 0, err

	}
	return s.segment.Create(ctx, name)
}

func (s *service) GetAllSegments(ctx context.Context) ([]model.SegmentInfo, error) {
	return s.segment.GetAll(ctx)
}
