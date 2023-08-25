package usersegments

import (
	"context"
	"testing"
	"time"

	history "github.com/VrMolodyakov/segment-api/internal/domain/history/model"
	segment "github.com/VrMolodyakov/segment-api/internal/domain/segment/model"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
)

type mockClock struct {
	currentTime time.Time
}

func NewTestClock(currentTime time.Time) *mockClock {
	return &mockClock{
		currentTime: currentTime,
	}
}

func (tc *mockClock) Now() time.Time {
	return tc.currentTime
}

func (tc *mockClock) Since(t time.Time) time.Duration {
	return tc.currentTime.Sub(t)
}

func (tc *mockClock) Until(t time.Time) time.Duration {
	return t.Sub(tc.currentTime)
}

func TestUpdateUserSegments(t *testing.T) {
	ctx := context.Background()
	mockPSQLClient, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockPSQLClient.Close()

	testTime := time.Date(2023, 8, 25, 12, 0, 0, 0, time.UTC)
	clock := NewTestClock(testTime)
	repo := New(mockPSQLClient, clock)

	userID := int64(1)
	addSegments := []segment.Segment{
		{ID: 1, Name: "segment1", ExpiredAt: testTime},
		{ID: 2, Name: "segment2", ExpiredAt: testTime},
	}
	deleteSegmentNames := []string{"segment3", "segment4"}

	insertNames := []interface{}{"segment1", "segment2"}
	deleteNames := []interface{}{"segment3", "segment4"}

	historyRows := []interface{}{
		userID, int64(1), history.Added, testTime,
		userID, int64(2), history.Added, testTime,
		userID, int64(3), history.Deleted, testTime,
		userID, int64(4), history.Deleted, testTime,
	}
	insertRecors := []interface{}{userID, int64(1), testTime, userID, int64(2), testTime}

	insertRows := mockPSQLClient.
		NewRows([]string{"segment_id", "segment_name"}).
		AddRow(int64(1), "segment1").
		AddRow(int64(2), "segment2")

	deleteRows := mockPSQLClient.
		NewRows([]string{"segment_id"}).
		AddRow(int64(3)).
		AddRow(int64(4))

	tests := []struct {
		title       string
		isError     bool
		expectedErr error
		mockCall    func()
	}{
		{
			title: "Should successfully insert 2 segment and delete 2 segment",
			mockCall: func() {

				mockPSQLClient.
					ExpectBegin()
				mockPSQLClient.
					ExpectQuery("SELECT segment_id, segment_name FROM segments WHERE segment_name IN ").
					WithArgs(insertNames...).
					WillReturnRows(insertRows)
				mockPSQLClient.
					ExpectQuery("SELECT segment_id FROM segments WHERE segment_name IN ").
					WithArgs(deleteNames...).
					WillReturnRows(deleteRows)
				mockPSQLClient.
					ExpectExec("DELETE FROM user_segments WHERE ").
					WithArgs(int64(1), int64(3), int64(4)).
					WillReturnResult(pgxmock.NewResult("DELETE", 2))
				mockPSQLClient.
					ExpectExec("INSERT INTO user_segments").
					WithArgs(insertRecors...).
					WillReturnResult(pgxmock.NewResult("INSERT", 2))
				mockPSQLClient.
					ExpectExec("INSERT INTO segment_history").
					WithArgs(historyRows...).
					WillReturnResult(pgxmock.NewResult("INSERT", 4))
				mockPSQLClient.
					ExpectCommit()
			},
			isError: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			err := repo.UpdateUserSegments(ctx, userID, addSegments, deleteSegmentNames)
			if test.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
