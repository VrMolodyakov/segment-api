package service

import (
	"context"
	"errors"

	"github.com/VrMolodyakov/segment-api/internal/domain/participation/model"
	segment "github.com/VrMolodyakov/segment-api/internal/domain/segment/model"
	"github.com/VrMolodyakov/segment-api/pkg/logging"
)

var ErrSegmentAlreadyAssigned = errors.New("segment already assigned")

type ParticipationRepository interface {
	UpdateUserSegments(ctx context.Context, userID int64, addSegments []segment.Segment, deleteSegments []string) error
	DeleteSegment(ctx context.Context, name string) error
	GetUserSegments(ctx context.Context, userID int64) ([]model.Participation, error)
}

type service struct {
	logger        logging.Logger
	participation ParticipationRepository
}

func New(participation ParticipationRepository, logger logging.Logger) *service {
	return &service{
		participation: participation,
		logger:        logger,
	}
}

func (s *service) UpdateUserParticipation(ctx context.Context, userID int64, addSegments []segment.Segment, deleteSegments []string) error {
	s.logger.Debugf("try to update user = %d segments %v delete segments %v", addSegments, deleteSegments)
	return s.participation.UpdateUserSegments(ctx, userID, addSegments, deleteSegments)
}

func (s *service) DeleteParticipation(ctx context.Context, segmentName string) error {
	s.logger.Debugf("try to delete %s segment", segmentName)
	return s.participation.DeleteSegment(ctx, segmentName)
}

func (s *service) GetParticipation(ctx context.Context, userID int64) ([]model.Participation, error) {
	s.logger.Debugf("try to get user %d segments", userID)
	return s.participation.GetUserSegments(ctx, userID)
}
