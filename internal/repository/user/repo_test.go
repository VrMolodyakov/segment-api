package user

import (
	"context"
	"errors"
	"testing"

	"github.com/VrMolodyakov/segment-api/internal/domain/user"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	mockPSQLClient, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockPSQLClient.Close()
	repo := New(mockPSQLClient)

	userID := int64(1)
	got := user.User{
		ID:        userID,
		FirstName: "Arnold",
		LastName:  "Jones",
		Email:     "t2000@mail.ru",
	}

	type args struct {
		userID int64
	}

	tests := []struct {
		title    string
		args     args
		expected user.User
		isError  bool
		mockCall func()
	}{
		{
			title: "Should successfully get the user",
			args: args{
				userID: userID,
			},
			isError: false,
			mockCall: func() {
				rows := pgxmock.NewRows([]string{"user_id", "first_name", "last_name", "email"}).
					AddRow(got.ID, got.FirstName, got.LastName, got.Email)

				mockPSQLClient.ExpectQuery("SELECT user_id, first_name, last_name, email FROM users").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expected: got,
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
			expected: user.User{},
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
			expected: user.User{},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			got, err := repo.Get(ctx, test.args.userID)
			if test.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, got)

		})
	}

}
