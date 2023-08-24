package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/VrMolodyakov/segment-api/internal/domain/user/model"
	"github.com/VrMolodyakov/segment-api/internal/domain/user/service"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	mockPSQLClient, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockPSQLClient.Close()
	repo := New(mockPSQLClient)

	type args struct {
		user model.User
	}

	userID := 1

	newUser := model.User{
		FirstName: "Arnold",
		LastName:  "Jones",
		Email:     "t2000@mail.ru",
	}

	tests := []struct {
		title       string
		args        args
		expected    int
		expectError bool
		mockCall    func()
	}{
		{
			title: "Should successfully insert a new user",
			args: args{
				user: newUser,
			},
			expectError: false,
			mockCall: func() {
				rows := pgxmock.NewRows([]string{"user_id"}).AddRow(userID)
				mockPSQLClient.
					ExpectQuery("INSERT INTO users").
					WithArgs(newUser.FirstName, newUser.LastName, newUser.Email).
					WillReturnRows(rows)
			},
			expected: userID,
		},
		{
			title: "Database internal error",
			args: args{
				user: newUser,
			},
			expectError: true,
			mockCall: func() {
				mockPSQLClient.
					ExpectQuery("INSERT INTO users").
					WithArgs(newUser.FirstName, newUser.LastName, newUser.Email).
					WillReturnError(errors.New("internal database error"))
			},
			expected: 0,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			got, err := repo.Create(ctx, test.args.user)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, got)

		})
	}

}

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	mockPSQLClient, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockPSQLClient.Close()
	repo := New(mockPSQLClient)

	userID := 1
	user := model.User{
		ID:        userID,
		FirstName: "Arnold",
		LastName:  "Jones",
		Email:     "t2000@mail.ru",
	}

	type args struct {
		userID int
	}

	tests := []struct {
		title         string
		args          args
		expected      model.User
		isError       bool
		expectedError error
		mockCall      func()
	}{
		{
			title: "Should successfully get the user",
			args: args{
				userID: userID,
			},
			isError: false,
			mockCall: func() {
				rows := pgxmock.NewRows([]string{"user_id", "first_name", "last_name", "email"}).
					AddRow(user.ID, user.FirstName, user.LastName, user.Email)

				mockPSQLClient.ExpectQuery("SELECT user_id, first_name, last_name, email FROM users").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expected: user,
		},
		{
			title: "Database internal error",
			args: args{
				userID: userID,
			},
			isError: true,
			mockCall: func() {
				mockPSQLClient.ExpectQuery("SELECT user_id, first_name, last_name, email FROM users").
					WithArgs(userID).
					WillReturnError(errors.New("internal database error"))
			},
			expected:      model.User{},
			expectedError: errors.New("internal database error"),
		},
		{
			title: "User not found",
			args: args{
				userID: userID,
			},
			isError: true,
			mockCall: func() {
				mockPSQLClient.ExpectQuery("SELECT user_id, first_name, last_name, email FROM users").
					WithArgs(userID).
					WillReturnError(pgx.ErrNoRows)
			},
			expected:      model.User{},
			expectedError: service.ErrUserNotFound,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			got, err := repo.Get(ctx, test.args.userID)
			if test.isError {
				assert.Error(t, err)
				assert.Equal(t, err, test.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, got)

		})
	}

}
