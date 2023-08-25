package usersegments

import (
	"context"
	"errors"
	"testing"
	"time"

	segment "github.com/VrMolodyakov/segment-api/internal/domain/segment/model"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
)

func TestUpdateUserSegments(t *testing.T) {
	ctx := context.Background()
	mockPSQLClient, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockPSQLClient.Close()
	repo := New(mockPSQLClient)

	userID := 1
	addSegments := []segment.Segment{
		{ID: 1, Name: "segment1", ExpiredAt: time.Now().Add(time.Hour * 24)},
	}
	deleteSegments := []string{"segment2"}

	tests := []struct {
		title       string
		isError     bool
		expectedErr error
		mockCall    func()
	}{
		{
			title: "Should successfully update user segments",
			mockCall: func() {
				mockPSQLClient.
					ExpectBegin()
				mockPSQLClient.
					ExpectCommit()
				mockPSQLClient.
					ExpectQuery("SELECT segment_id, segment_name FROM segments").
					WithArgs("segment1").
					WillReturnRows(pgxmock.NewRows([]string{"segment_id", "segment_name"}).
						AddRow(addSegments[0].ID, addSegments[0].Name))
				mockPSQLClient.
					ExpectQuery("SELECT segment_id FROM segments WHERE segment_name = ANY($1)").
					WithArgs(addSegments).
					WillReturnRows(pgxmock.NewRows([]string{"segment_id"}).AddRow(2))
				mockPSQLClient.
					ExpectExec("DELETE FROM user_segments").
					WithArgs(userID, 2).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
				mockPSQLClient.
					ExpectExec("INSERT INTO user_segments").
					WithArgs(userID, addSegments[0].ID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mockPSQLClient.
					ExpectExec("INSERT INTO history").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			isError: false,
		},
		{
			title: "Database internal error",
			mockCall: func() {
				mockPSQLClient.
					ExpectBegin()
				mockPSQLClient.
					ExpectRollback()
				mockPSQLClient.
					ExpectQuery("SELECT segment_id, segment_name FROM segments").
					WithArgs("segment1").
					WillReturnError(errors.New("internal database error"))
			},
			isError:     true,
			expectedErr: errors.New("internal database error"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			err := repo.UpdateUserSegments(ctx, userID, addSegments, deleteSegments)
			if test.isError {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
