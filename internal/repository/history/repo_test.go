package history

import (
	"context"
	"errors"
	"testing"
	"time"

	history "github.com/VrMolodyakov/segment-api/internal/domain/history/model"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetUserSegments(t *testing.T) {
	ctx := context.Background()
	mockClient, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockClient.Close()
	testTime := time.Date(2023, 8, 25, 12, 0, 0, 0, time.UTC)
	repo := New(mockClient)

	userID := int64(1)
	year, month := 2013, 11
	historyRecords := []history.History{
		{ID: int64(1), UserID: userID, Segment: "segment1", Operation: "Added", Time: testTime},
		{ID: int64(2), UserID: userID, Segment: "segment1", Operation: "Deleted", Time: testTime},
	}

	type args struct {
		year  int
		month int
	}

	tests := []struct {
		title       string
		isError     bool
		expected    []history.History
		expectedErr error
		args        args
		mockCall    func()
	}{
		{
			title: "Should successfully retrieve user segments history",
			mockCall: func() {
				rows := pgxmock.NewRows([]string{"history_id", "user_id", "segment_name", "operation", "operation_timestamp"}).
					AddRow(historyRecords[0].ID, historyRecords[0].UserID, historyRecords[0].Segment, historyRecords[0].Operation, historyRecords[0].Time).
					AddRow(historyRecords[1].ID, historyRecords[1].UserID, historyRecords[1].Segment, historyRecords[1].Operation, historyRecords[1].Time)
				mockClient.
					ExpectQuery("SELECT user_id, segment_name, operation, operation_timestamp FROM segment_history JOIN segments").
					WithArgs(year, month).
					WillReturnRows(rows)
			},
			args:     args{year: year, month: month},
			isError:  false,
			expected: historyRecords,
		},
		{
			title: "Database internal error",
			mockCall: func() {
				mockClient.
					ExpectQuery("SELECT user_id, segment_name, operation, operation_timestamp FROM segment_history JOIN segments").
					WithArgs(year, month).
					WillReturnError(errors.New("internal database error"))
			},
			args:        args{year: year, month: month},
			isError:     true,
			expected:    nil,
			expectedErr: errors.New("internal database error"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			result, err := repo.Get(ctx, test.args.year, test.args.month)
			if test.isError {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)

			}
			assert.Equal(t, test.expected, result)
		})
	}
}
