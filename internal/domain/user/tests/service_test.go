package user

import (
	"context"
	"errors"
	"testing"

	"github.com/VrMolodyakov/segment-api/internal/domain/user"
	"github.com/VrMolodyakov/segment-api/internal/domain/user/mocks"
	"github.com/VrMolodyakov/segment-api/pkg/logging"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockLogger, err := logging.MockLogger()
	assert.NoError(t, err)
	userService := user.New(mockRepo, mockLogger)
	ctx := context.Background()
	type mockCall func()

	type args struct {
		id int64
	}

	userID := int64(1)

	insert := user.User{
		ID: userID,
	}

	testCases := []struct {
		title    string
		mockCall mockCall
		args     args
		expected user.User
		isError  bool
	}{
		{
			title: "Successful getting user",
			mockCall: func() {
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(insert, nil)
			},
			args: args{
				userID,
			},
			expected: insert,
		},
		{
			title: "User not found and should return ErrUserNotFound",
			mockCall: func() {
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(user.User{}, user.ErrUserNotFound)
			},
			args: args{
				userID,
			},
			isError:  true,
			expected: user.User{},
		},
		{
			title: "DB error",
			mockCall: func() {
				mockRepo.EXPECT().Get(gomock.Any(), gomock.Any()).Return(user.User{}, errors.New("db internal error"))
			},
			args: args{
				userID,
			},
			isError:  true,
			expected: user.User{},
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
