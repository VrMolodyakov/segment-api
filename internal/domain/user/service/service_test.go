package service

import (
	"context"
	"errors"
	"testing"

	"github.com/VrMolodyakov/segment-api/internal/domain/user/model"
	"github.com/VrMolodyakov/segment-api/internal/domain/user/service/mocks"
	"github.com/VrMolodyakov/segment-api/pkg/logging"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSaveUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockLogger, err := logging.MockLogger()
	assert.NoError(t, err)
	userService := New(mockRepo, mockLogger)
	ctx := context.Background()
	type mockCall func()

	type args struct {
		user model.User
	}
	userID := int64(1)
	testCases := []struct {
		title    string
		mockCall mockCall
		args     args
		expected int64
		isError  bool
	}{
		{
			title: "Successful user creation",
			mockCall: func() {
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(userID, nil)
			},
			args: args{
				model.User{FirstName: "abc", LastName: "efg", Email: "some@mail.com"},
			},
			expected: userID,
		},
		{
			title: "User already exists and should return ErrUserAlreadyExist",
			mockCall: func() {
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(0), ErrUserAlreadyExist)
			},
			args: args{
				model.User{FirstName: "abc", LastName: "efg", Email: "some@mail.com"},
			},
			isError:  true,
			expected: int64(0),
		},
		{
			title: "DB error",
			mockCall: func() {
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("db internal error"))
			},
			args: args{
				model.User{FirstName: "abc", LastName: "efg", Email: "some@mail.com"},
			},
			isError:  true,
			expected: int64(0),
		},
	}
	for _, test := range testCases {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			got, err := userService.CreateUser(ctx, test.args.user)
			if test.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestGetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockLogger, err := logging.MockLogger()
	assert.NoError(t, err)
	userService := New(mockRepo, mockLogger)
	ctx := context.Background()
	type mockCall func()

	type args struct {
		id int64
	}

	userID := int64(1)

	user := model.User{
		ID: userID,
	}

	testCases := []struct {
		title    string
		mockCall mockCall
		args     args
		expected model.User
		isError  bool
	}{
		{
			title: "Successful getting user",
			mockCall: func() {
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(user, nil)
			},
			args: args{
				userID,
			},
			expected: user,
		},
		{
			title: "User not found and should return ErrUserNotFound",
			mockCall: func() {
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(model.User{}, ErrUserNotFound)
			},
			args: args{
				userID,
			},
			isError:  true,
			expected: model.User{},
		},
		{
			title: "DB error",
			mockCall: func() {
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(model.User{}, errors.New("db internal error"))
			},
			args: args{
				userID,
			},
			isError:  true,
			expected: model.User{},
		},
	}
	for _, test := range testCases {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			got, err := userService.GetUser(ctx, test.args.id)
			if test.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, got)
		})
	}
}
