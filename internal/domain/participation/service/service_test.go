package service

import (
	"context"
	"errors"
	"testing"
	"time"

	participation "github.com/VrMolodyakov/segment-api/internal/domain/participation/model"
	"github.com/VrMolodyakov/segment-api/internal/domain/participation/service/mocks"
	"github.com/VrMolodyakov/segment-api/internal/domain/segment/model"
	segment "github.com/VrMolodyakov/segment-api/internal/domain/segment/service"
	"github.com/VrMolodyakov/segment-api/pkg/logging"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateUserParticipation(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockParticipationRepository(ctrl)
	defer ctrl.Finish()
	mockLogger, err := logging.MockLogger()
	assert.NoError(t, err)
	participationService := New(mockRepo, mockLogger)
	ctx := context.Background()
	type mockCall func()

	type args struct {
		userID int64
		add    []model.Segment
		delete []string
	}

	userID := int64(1)
	add := []model.Segment{
		{Name: "s-1"},
		{Name: "s-2"},
	}
	delete := []string{
		"s-3",
		"s-4",
	}

	testCases := []struct {
		title     string
		mockCall  mockCall
		args      args
		expectErr error
		isError   bool
	}{
		{
			title: "Successful update user segments",
			mockCall: func() {
				mockRepo.EXPECT().UpdateUserSegments(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			args: args{
				add:    add,
				delete: delete,
				userID: userID,
			},
		},
		{
			title: "Segment already assigned and should return ErrSegmentAlreadyAssigned",
			mockCall: func() {
				mockRepo.EXPECT().UpdateUserSegments(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(ErrSegmentAlreadyAssigned)
			},
			args: args{
				add:    add,
				delete: delete,
				userID: userID,
			},
			isError:   true,
			expectErr: ErrSegmentAlreadyAssigned,
		},
	}
	for _, test := range testCases {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			err := participationService.UpdateUserParticipation(ctx, test.args.userID, test.args.add, test.args.delete)
			if test.isError {
				assert.Error(t, err)
				assert.Equal(t, test.expectErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteParticipation(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockParticipationRepository(ctrl)
	defer ctrl.Finish()
	mockLogger, err := logging.MockLogger()
	assert.NoError(t, err)
	participationService := New(mockRepo, mockLogger)
	ctx := context.Background()
	type mockCall func()

	type args struct {
		name string
	}

	testCases := []struct {
		title     string
		mockCall  mockCall
		args      args
		expectErr error
		isError   bool
	}{
		{
			title: "Successful segment deletion",
			mockCall: func() {
				mockRepo.EXPECT().DeleteSegment(gomock.Any(), gomock.Any()).Return(nil)
			},
			args: args{
				name: "seg-1",
			},
		},
		{
			title: "Segment not found and should return ErrSegmentNotFound",
			mockCall: func() {
				mockRepo.EXPECT().DeleteSegment(gomock.Any(), gomock.Any()).Return(segment.ErrSegmentNotFound)
			},
			args: args{
				name: "seg-1",
			},
			isError:   true,
			expectErr: segment.ErrSegmentNotFound,
		},
	}
	for _, test := range testCases {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			err := participationService.DeleteParticipation(ctx, test.args.name)
			if test.isError {
				assert.Error(t, err)
				assert.Equal(t, test.expectErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUserSegments(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockParticipationRepository(ctrl)
	defer ctrl.Finish()
	mockLogger, err := logging.MockLogger()
	assert.NoError(t, err)
	participationService := New(mockRepo, mockLogger)
	ctx := context.Background()
	type mockCall func()

	userID := int64(1)
	type args struct {
		userID int64
	}

	parts := []participation.Participation{
		{UserID: 1, SegmentName: "seg-1", ExpiredAt: time.Date(2023, 8, 25, 12, 0, 0, 0, time.UTC)},
		{UserID: 2, SegmentName: "seg-2", ExpiredAt: time.Date(2023, 8, 25, 12, 0, 0, 0, time.UTC)},
	}

	testCases := []struct {
		title    string
		mockCall mockCall
		args     args
		expected []participation.Participation
		isError  bool
	}{
		{
			title: "Successful get user segment participation",
			mockCall: func() {
				mockRepo.EXPECT().GetUserSegments(gomock.Any(), gomock.Any()).Return(parts, nil)
			},
			args: args{
				userID: userID,
			},
			expected: parts,
		},
		{
			title: "DB error",
			mockCall: func() {
				mockRepo.EXPECT().GetUserSegments(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
			},
			args: args{
				userID: userID,
			},
			expected: nil,
		},
	}
	for _, test := range testCases {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			got, err := participationService.GetParticipation(ctx, test.args.userID)
			if test.isError {
				assert.Error(t, err)
			} else {
				assert.Equal(t, test.expected, got)
			}
		})
	}
}
