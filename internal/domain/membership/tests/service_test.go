package membership

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/VrMolodyakov/segment-api/internal/domain/membership"
	"github.com/VrMolodyakov/segment-api/internal/domain/membership/mocks"
	"github.com/VrMolodyakov/segment-api/internal/domain/segment"
	"github.com/VrMolodyakov/segment-api/internal/domain/user"
	"github.com/VrMolodyakov/segment-api/pkg/logging"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type mockRandom struct{}

func (m *mockRandom) Next() int { return 0 }

func TestUpdateUserMembership(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockMembershipRepository(ctrl)
	defer ctrl.Finish()
	mockLogger, err := logging.MockLogger()
	assert.NoError(t, err)
	mockCache := mocks.NewMockCache(ctrl)
	membershipService := membership.New(mockRepo, mockCache, 1*time.Minute, &mockRandom{}, mockLogger)
	ctx := context.Background()
	type mockCall func()

	type args struct {
		userID int64
		add    []segment.Segment
		delete []string
	}

	userID := int64(1)
	add := []segment.Segment{
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
			title: "Segment already assigned error",
			mockCall: func() {
				mockRepo.EXPECT().UpdateUserSegments(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(membership.ErrSegmentAlreadyAssigned)
			},
			args: args{
				add:    add,
				delete: delete,
				userID: userID,
			},
			isError:   true,
			expectErr: membership.ErrSegmentAlreadyAssigned,
		},
		{
			title: "Segment already assigned error",
			mockCall: func() {
				mockRepo.EXPECT().UpdateUserSegments(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(user.ErrUserNotFound)
			},
			args: args{
				add:    add,
				delete: delete,
				userID: userID,
			},
			isError:   true,
			expectErr: user.ErrUserNotFound,
		},
		{
			title: "Segment not found error",
			mockCall: func() {
				mockRepo.EXPECT().UpdateUserSegments(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(segment.ErrSegmentNotFound)
			},
			args: args{
				add:    add,
				delete: delete,
				userID: userID,
			},
			isError:   true,
			expectErr: segment.ErrSegmentNotFound,
		},
		{
			title: "Empty data error",
			mockCall: func() {
			},
			args: args{
				add:    nil,
				delete: nil,
				userID: userID,
			},
			isError:   true,
			expectErr: membership.ErrEmptyData,
		},
		{
			title: "incorrect data error, the same elements",
			mockCall: func() {
			},
			args: args{
				add:    []segment.Segment{{Name: "seg-2"}, {Name: "seg-9"}},
				delete: []string{"seg-1", "seg-2", "seg-10"},
				userID: userID,
			},
			isError:   true,
			expectErr: membership.ErrIncorrectData,
		},
	}
	for _, test := range testCases {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			err := membershipService.UpdateUserMembership(ctx, test.args.userID, test.args.add, test.args.delete)
			if test.isError {
				assert.Error(t, err)
				assert.Equal(t, test.expectErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteMembership(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockMembershipRepository(ctrl)
	defer ctrl.Finish()
	mockLogger, err := logging.MockLogger()
	assert.NoError(t, err)
	mockCache := mocks.NewMockCache(ctrl)
	membershipService := membership.New(mockRepo, mockCache, 1*time.Minute, &mockRandom{}, mockLogger)
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
			err := membershipService.DeleteMembership(ctx, test.args.name)
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
	mockRepo := mocks.NewMockMembershipRepository(ctrl)
	defer ctrl.Finish()
	mockLogger, err := logging.MockLogger()
	assert.NoError(t, err)
	mockCache := mocks.NewMockCache(ctrl)
	membershipService := membership.New(mockRepo, mockCache, 1*time.Minute, &mockRandom{}, mockLogger)
	ctx := context.Background()
	type mockCall func()

	userID := int64(1)
	type args struct {
		userID int64
	}

	info := []membership.MembershipInfo{
		{UserID: 1, SegmentName: "seg-1", ExpiredAt: time.Date(2023, 8, 25, 12, 0, 0, 0, time.UTC)},
		{UserID: 2, SegmentName: "seg-2", ExpiredAt: time.Date(2023, 8, 25, 12, 0, 0, 0, time.UTC)},
	}

	testCases := []struct {
		title    string
		mockCall mockCall
		args     args
		expected []membership.MembershipInfo
		isError  bool
	}{
		{
			title: "Not found in cache and successfully retrieves from repo",
			mockCall: func() {
				mockCache.EXPECT().Get(gomock.Any()).Return(nil, false)
				mockRepo.EXPECT().GetUserSegments(gomock.Any(), gomock.Any()).Return(info, nil)
				mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any())
			},
			args: args{
				userID: userID,
			},
			expected: info,
		},
		{
			title: "Found in the cache and return the resulting value",
			mockCall: func() {
				mockCache.EXPECT().Get(gomock.Any()).Return(info, true)
			},
			args: args{
				userID: userID,
			},
			expected: info,
		},
		{
			title: "Not found in cache and could not get from repository",
			mockCall: func() {
				mockCache.EXPECT().Get(gomock.Any()).Return(nil, false)
				mockRepo.EXPECT().GetUserSegments(gomock.Any(), gomock.Any()).Return(nil, errors.New("couldn't get data"))
			},
			args: args{
				userID: userID,
			},
			isError:  true,
			expected: nil,
		},
	}
	for _, test := range testCases {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			got, err := membershipService.GetUserMembership(ctx, test.args.userID)
			if test.isError {
				assert.Error(t, err)
			} else {
				assert.Equal(t, test.expected, got)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockMembershipRepository(ctrl)
	defer ctrl.Finish()
	mockLogger, err := logging.MockLogger()
	assert.NoError(t, err)
	mockCache := mocks.NewMockCache(ctrl)
	membershipService := membership.New(mockRepo, mockCache, 1*time.Minute, &mockRandom{}, mockLogger)
	ctx := context.Background()
	type mockCall func()

	userID := int64(1)
	emptyID := int64(0)
	type args struct {
		user user.User
	}

	testCases := []struct {
		title         string
		mockCall      mockCall
		args          args
		expected      int64
		expectedError error
		isError       bool
	}{
		{
			title: "Successful user creation",
			mockCall: func() {
				mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(userID, nil)
			},
			args: args{
				user: user.User{
					Email: "email@email.com",
				},
			},
			expected: userID,
		},
		{
			title: "Invalid email error",
			mockCall: func() {
				mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(emptyID, user.ErrInvalidEmail)
			},
			args: args{
				user: user.User{
					Email: "email@email.com",
				},
			},
			expectedError: user.ErrInvalidEmail,
			expected:      emptyID,
			isError:       true,
		},
		{
			title: "User already exists error",
			mockCall: func() {
				mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(emptyID, user.ErrUserAlreadyExist)
			},
			args: args{
				user: user.User{
					Email: "email@email.com",
				},
			},
			expectedError: user.ErrUserAlreadyExist,
			expected:      emptyID,
			isError:       true,
		},
		{
			title: "Error while inserting new user",
			mockCall: func() {
				mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(emptyID, errors.New("error"))
			},
			args: args{
				user: user.User{
					Email: "email@email.com",
				},
			},
			expectedError: errors.New("error"),
			expected:      emptyID,
			isError:       true,
		},
	}
	for _, test := range testCases {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			got, err := membershipService.CreateUser(ctx, test.args.user)
			if test.isError {
				assert.Error(t, err)
			} else {
				assert.Equal(t, test.expected, got)
			}
		})
	}
}
