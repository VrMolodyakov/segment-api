package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/VrMolodyakov/segment-api/internal/domain/history"
	"github.com/VrMolodyakov/segment-api/internal/domain/history/mocks"
	"github.com/VrMolodyakov/segment-api/pkg/logging"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPrepareHistoryData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger, err := logging.MockLogger()
	assert.NoError(t, err)
	mockCache := mocks.NewMockCache(ctrl)
	mockRepo := mocks.NewMockHistoryRepository(ctrl)
	historyService := history.New(mockRepo, mockCache, 1*time.Hour, mockLogger)
	ctx := context.Background()

	type mockCall func()
	type args struct {
		date history.Date
	}
	histories := []history.History{{UserID: 1}, {UserID: 2}}
	testCases := []struct {
		title    string
		mockCall mockCall
		args     args
		isError  bool
	}{
		{
			title: "Not found in the cache and successfully retrieved from the repo",
			mockCall: func() {
				mockCache.EXPECT().Get(gomock.Any()).Return(nil, false)
				mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(histories, nil)

			},
			args: args{
				history.Date{
					Year:  2023,
					Month: 8,
				},
			},
		},
		{
			title: "Successfully found in the cache",
			mockCall: func() {
				mockCache.EXPECT().Get(gomock.Any()).Return(histories, true)
			},
			args: args{
				history.Date{
					Year:  2023,
					Month: 8,
				},
			},
		},
		{
			title: "Not found in the cache and error while retrieving from the repo",
			mockCall: func() {
				mockCache.EXPECT().Get(gomock.Any()).Return(nil, false)
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("repo error"))

			},
			args: args{
				history.Date{
					Year:  2023,
					Month: 8,
				},
			},
			isError: true,
		},
		{
			title: "Validation error, incorrect year",
			mockCall: func() {
			},
			args: args{
				history.Date{
					Year:  1988,
					Month: 7,
				},
			},
			isError: true,
		},
		{
			title: "Validation error, incorrect month",
			mockCall: func() {
			},
			args: args{
				history.Date{
					Year:  2023,
					Month: 13,
				},
			},
			isError: true,
		},
	}
	for _, test := range testCases {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			err := historyService.PrepareHistoryData(ctx, test.args.date)
			if test.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUsersHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger, err := logging.MockLogger()
	assert.NoError(t, err)
	mockCache := mocks.NewMockCache(ctrl)
	mockRepo := mocks.NewMockHistoryRepository(ctrl)
	historyService := history.New(mockRepo, mockCache, 1*time.Hour, mockLogger)
	ctx := context.Background()

	type mockCall func()
	type args struct {
		date history.Date
	}
	histories := []history.History{{UserID: 1}, {UserID: 2}}
	testCases := []struct {
		title    string
		mockCall mockCall
		args     args
		expected []history.History
		isError  bool
	}{
		{
			title: "successfully found the data in the cache",
			mockCall: func() {
				mockCache.EXPECT().Get(gomock.Any()).Return(histories, true)

			},
			args: args{
				history.Date{
					Year:  2023,
					Month: 8,
				},
			},
			expected: histories,
		},
		{
			title: "Not found in the cache and return error",
			mockCall: func() {
				mockCache.EXPECT().Get(gomock.Any()).Return(nil, false)
			},
			args: args{
				history.Date{
					Year:  2023,
					Month: 8,
				},
			},
			isError:  true,
			expected: nil,
		},
		{
			title: "Validation error, incorrect year",
			mockCall: func() {
			},
			args: args{
				history.Date{
					Year:  1988,
					Month: 7,
				},
			},
			isError:  true,
			expected: nil,
		},
		{
			title: "Validation error, incorrect month",
			mockCall: func() {
			},
			args: args{
				history.Date{
					Year:  2023,
					Month: 13,
				},
			},
			isError:  true,
			expected: nil,
		},
	}
	for _, test := range testCases {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			got, err := historyService.GetUsersHistory(ctx, test.args.date)
			if test.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, got)
		})
	}
}
