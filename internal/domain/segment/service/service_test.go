package service

import (
	"context"
	"errors"
	"testing"

	"github.com/VrMolodyakov/segment-api/internal/domain/segment/model"
	"github.com/VrMolodyakov/segment-api/internal/domain/segment/service/mocks"
	"github.com/VrMolodyakov/segment-api/pkg/logging"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSaveSegment(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockSegmentRepository(ctrl)
	defer ctrl.Finish()
	mockLogger, err := logging.MockLogger()
	assert.NoError(t, err)
	segmentService := New(mockRepo, mockLogger)
	ctx := context.Background()
	type mockCall func()

	type args struct {
		segment string
	}
	segmentID := int64(1)
	emptyID := int64(0)
	testCases := []struct {
		title    string
		mockCall mockCall
		args     args
		expected int64
		isError  bool
	}{
		{
			title: "Successful segment creation",
			mockCall: func() {
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(model.SegmentInfo{}, ErrSegmentNotFound)
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(segmentID, nil)
			},
			args: args{
				"segment1",
			},
			expected: segmentID,
		},
		{
			title: "Couldn't get sigment info should return error",
			mockCall: func() {
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(model.SegmentInfo{}, errors.New("db internal error"))
			},
			args: args{
				"segment1",
			},
			isError:  true,
			expected: emptyID,
		},
		{
			title: "Segment already exists should return error",
			mockCall: func() {
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(model.SegmentInfo{Name: "segment1"}, nil)
			},
			args: args{
				"segment1",
			},
			isError:  true,
			expected: emptyID,
		},
	}
	for _, test := range testCases {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			got, err := segmentService.CreateSegment(ctx, test.args.segment)
			if test.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestGetAllSegments(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockSegmentRepository(ctrl)
	defer ctrl.Finish()
	mockLogger, err := logging.MockLogger()
	assert.NoError(t, err)
	segmentService := New(mockRepo, mockLogger)
	ctx := context.Background()
	type mockCall func()

	s1 := model.SegmentInfo{Name: "segment1"}
	s2 := model.SegmentInfo{Name: "segment2"}

	testCases := []struct {
		title    string
		mockCall mockCall
		expected []model.SegmentInfo
		isError  bool
	}{
		{
			title: "Successful getting segments",
			mockCall: func() {
				mockRepo.EXPECT().GetAll(gomock.All()).Return([]model.SegmentInfo{s1, s2}, nil)
			},
			expected: []model.SegmentInfo{s1, s2},
		},
		{
			title: "Couldn't get segments should return error",
			mockCall: func() {
				mockRepo.EXPECT().GetAll(gomock.All()).Return(nil, errors.New("db error"))
			},
			isError:  true,
			expected: nil,
		},
	}
	for _, test := range testCases {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			got, err := segmentService.GetAllSegments(ctx)
			if test.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, got)
		})
	}
}
