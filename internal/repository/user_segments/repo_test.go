package usersegments

import (
	"context"
	"errors"
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
	mockClient, err := pgxmock.NewPool()
	if err != nil {
		t.Error(err)
	}
	defer mockClient.Close()

	testTime := time.Date(2023, 8, 25, 12, 0, 0, 0, time.UTC)
	clock := NewTestClock(testTime)
	repo := New(mockClient, clock)

	userID := int64(1)
	insertID1, insertID2, deleteID1, deleteID2 := int64(1), int64(2), int64(3), int64(4)

	type args struct {
		addSegments        []segment.Segment
		deleteSegmentNames []string
		userID             int64
	}

	tests := []struct {
		title       string
		isError     bool
		expectedErr error
		args        args
		mockCall    func()
	}{
		{
			title: "Should successfully insert 2 segment and delete 2 segment",
			mockCall: func() {
				insertNames := []interface{}{"segment1", "segment2"}
				deleteNames := []interface{}{"segment3", "segment4"}
				insertRecors := []interface{}{userID, insertID1, testTime, userID, insertID2, testTime}
				deleteRecors := []interface{}{userID, deleteID1, deleteID2}
				insertRows := pgxmock.
					NewRows([]string{"segment_id", "segment_name"}).
					AddRow(insertID1, "segment1").
					AddRow(insertID2, "segment2")
				deleteRows := pgxmock.
					NewRows([]string{"segment_id"}).
					AddRow(deleteID1).
					AddRow(deleteID2)
				historyRows := []interface{}{
					userID, insertID1, history.Added, testTime,
					userID, insertID2, history.Added, testTime,
					userID, deleteID1, history.Deleted, testTime,
					userID, deleteID2, history.Deleted, testTime,
				}

				mockClient.
					ExpectBegin()
				mockClient.
					ExpectQuery("SELECT segment_id, segment_name FROM segments WHERE segment_name IN ").
					WithArgs(insertNames...).
					WillReturnRows(insertRows)
				mockClient.
					ExpectExec("INSERT INTO user_segments").
					WithArgs(insertRecors...).
					WillReturnResult(pgxmock.NewResult("INSERT", 2))
				mockClient.
					ExpectQuery("SELECT segment_id FROM segments WHERE segment_name IN ").
					WithArgs(deleteNames...).
					WillReturnRows(deleteRows)
				mockClient.
					ExpectExec("DELETE FROM user_segments WHERE ").
					WithArgs(deleteRecors...).
					WillReturnResult(pgxmock.NewResult("DELETE", 2))
				mockClient.
					ExpectExec("INSERT INTO segment_history").
					WithArgs(historyRows...).
					WillReturnResult(pgxmock.NewResult("INSERT", 4))
				mockClient.
					ExpectCommit()
			},
			args: args{
				addSegments: []segment.Segment{
					{ID: 1, Name: "segment1", ExpiredAt: testTime},
					{ID: 2, Name: "segment2", ExpiredAt: testTime},
				},
				deleteSegmentNames: []string{"segment3", "segment4"},
				userID:             userID,
			},
			isError: false,
		},
		{

			title: "Couldn't find all the ids by the name to add and got an error",
			mockCall: func() {
				insertNames := []interface{}{"segment1", "segment2"}
				mockClient.
					ExpectBegin()
				mockClient.
					ExpectQuery("SELECT segment_id, segment_name FROM segments WHERE segment_name IN ").
					WithArgs(insertNames...).
					WillReturnError(errors.New("couldn't find some id"))
				mockClient.ExpectRollback()
			},
			args: args{
				addSegments: []segment.Segment{
					{ID: 1, Name: "segment1", ExpiredAt: testTime},
					{ID: 2, Name: "segment2", ExpiredAt: testTime},
				},
				deleteSegmentNames: []string{"segment3", "segment4"},
				userID:             userID,
			},
			isError:     true,
			expectedErr: errors.New("couldn't find some id"),
		},
		{
			title: "Couldn't find all the ids by the name to delete and got an error",
			mockCall: func() {
				insertNames := []interface{}{"segment1", "segment2"}
				deleteNames := []interface{}{"segment3", "segment4"}
				insertRecors := []interface{}{userID, insertID1, testTime, userID, insertID2, testTime}
				insertRows := pgxmock.
					NewRows([]string{"segment_id", "segment_name"}).
					AddRow(insertID1, "segment1").
					AddRow(insertID2, "segment2")

				mockClient.
					ExpectBegin()
				mockClient.
					ExpectQuery("SELECT segment_id, segment_name FROM segments WHERE segment_name IN ").
					WithArgs(insertNames...).
					WillReturnRows(insertRows)
				mockClient.
					ExpectExec("INSERT INTO user_segments").
					WithArgs(insertRecors...).
					WillReturnResult(pgxmock.NewResult("INSERT", 2))
				mockClient.
					ExpectQuery("SELECT segment_id FROM segments WHERE segment_name IN ").
					WithArgs(deleteNames...).
					WillReturnError(errors.New("couldn't find some id"))
				mockClient.ExpectRollback()
			},
			args: args{
				addSegments: []segment.Segment{
					{ID: 1, Name: "segment1", ExpiredAt: testTime},
					{ID: 2, Name: "segment2", ExpiredAt: testTime},
				},
				deleteSegmentNames: []string{"segment3", "segment4"},
				userID:             userID,
			},
			isError:     true,
			expectedErr: errors.New("couldn't find some id"),
		},
		{
			title: "Couldn't delete the necessary columns and got an error",
			mockCall: func() {
				insertNames := []interface{}{"segment1", "segment2"}
				deleteNames := []interface{}{"segment3", "segment4"}
				insertRecors := []interface{}{userID, insertID1, testTime, userID, insertID2, testTime}
				deleteRecors := []interface{}{userID, deleteID1, deleteID2}
				insertRows := pgxmock.
					NewRows([]string{"segment_id", "segment_name"}).
					AddRow(insertID1, "segment1").
					AddRow(insertID2, "segment2")
				deleteRows := pgxmock.
					NewRows([]string{"segment_id"}).
					AddRow(deleteID1).
					AddRow(deleteID2)

				mockClient.
					ExpectBegin()
				mockClient.
					ExpectQuery("SELECT segment_id, segment_name FROM segments WHERE segment_name IN ").
					WithArgs(insertNames...).
					WillReturnRows(insertRows)
				mockClient.
					ExpectExec("INSERT INTO user_segments").
					WithArgs(insertRecors...).
					WillReturnResult(pgxmock.NewResult("INSERT", 2))
				mockClient.
					ExpectQuery("SELECT segment_id FROM segments WHERE segment_name IN ").
					WithArgs(deleteNames...).
					WillReturnRows(deleteRows)
				mockClient.
					ExpectExec("DELETE FROM user_segments WHERE ").
					WithArgs(deleteRecors...).
					WillReturnError(errors.New("couldn't delete some id"))
				mockClient.
					ExpectRollback()
			},
			args: args{
				addSegments: []segment.Segment{
					{ID: 1, Name: "segment1", ExpiredAt: testTime},
					{ID: 2, Name: "segment2", ExpiredAt: testTime},
				},
				deleteSegmentNames: []string{"segment3", "segment4"},
				userID:             userID,
			},
			isError:     true,
			expectedErr: errors.New("couldn't delete some id"),
		},
		{
			title: "Couldn't insert the necessary columns and got an error",
			mockCall: func() {
				insertNames := []interface{}{"segment1", "segment2"}
				insertRecors := []interface{}{userID, insertID1, testTime, userID, insertID2, testTime}
				insertRows := pgxmock.
					NewRows([]string{"segment_id", "segment_name"}).
					AddRow(insertID1, "segment1").
					AddRow(insertID2, "segment2")
				mockClient.
					ExpectBegin()
				mockClient.
					ExpectQuery("SELECT segment_id, segment_name FROM segments WHERE segment_name IN ").
					WithArgs(insertNames...).
					WillReturnRows(insertRows)
				mockClient.
					ExpectExec("INSERT INTO user_segments").
					WithArgs(insertRecors...).
					WillReturnError(errors.New("couldn't insert some id"))
				mockClient.
					ExpectRollback()
			},
			args: args{
				addSegments: []segment.Segment{
					{ID: 1, Name: "segment1", ExpiredAt: testTime},
					{ID: 2, Name: "segment2", ExpiredAt: testTime},
				},
				deleteSegmentNames: []string{"segment3", "segment4"},
				userID:             userID,
			},
			isError:     true,
			expectedErr: errors.New("couldn't insert some id"),
		},
		{
			title: "Couldn't insert the necessary columns in segments history and got an error",
			mockCall: func() {
				insertNames := []interface{}{"segment1", "segment2"}
				deleteNames := []interface{}{"segment3", "segment4"}
				insertRecors := []interface{}{userID, insertID1, testTime, userID, insertID2, testTime}
				deleteRecors := []interface{}{userID, deleteID1, deleteID2}
				insertRows := pgxmock.
					NewRows([]string{"segment_id", "segment_name"}).
					AddRow(insertID1, "segment1").
					AddRow(insertID2, "segment2")
				deleteRows := pgxmock.
					NewRows([]string{"segment_id"}).
					AddRow(deleteID1).
					AddRow(deleteID2)
				historyRows := []interface{}{
					userID, insertID1, history.Added, testTime,
					userID, insertID2, history.Added, testTime,
					userID, deleteID1, history.Deleted, testTime,
					userID, deleteID2, history.Deleted, testTime,
				}

				mockClient.
					ExpectBegin()
				mockClient.
					ExpectQuery("SELECT segment_id, segment_name FROM segments WHERE segment_name IN ").
					WithArgs(insertNames...).
					WillReturnRows(insertRows)
				mockClient.
					ExpectExec("INSERT INTO user_segments").
					WithArgs(insertRecors...).
					WillReturnResult(pgxmock.NewResult("INSERT", 2))
				mockClient.
					ExpectQuery("SELECT segment_id FROM segments WHERE segment_name IN ").
					WithArgs(deleteNames...).
					WillReturnRows(deleteRows)
				mockClient.
					ExpectExec("DELETE FROM user_segments WHERE ").
					WithArgs(deleteRecors...).
					WillReturnResult(pgxmock.NewResult("DELETE", 2))
				mockClient.
					ExpectExec("INSERT INTO segment_history").
					WithArgs(historyRows...).
					WillReturnError(errors.New("couldn't insert some rows"))
				mockClient.
					ExpectRollback()
			},
			args: args{
				addSegments: []segment.Segment{
					{ID: 1, Name: "segment1", ExpiredAt: testTime},
					{ID: 2, Name: "segment2", ExpiredAt: testTime},
				},
				deleteSegmentNames: []string{"segment3", "segment4"},
				userID:             userID,
			},
			isError:     true,
			expectedErr: errors.New("couldn't insert some rows"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			err := repo.UpdateUserSegments(
				ctx,
				test.args.userID,
				test.args.addSegments,
				test.args.deleteSegmentNames,
			)
			if test.isError {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestGetUserSegments(t *testing.T) {
	ctx := context.Background()
	mockClient, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockClient.Close()
	testTime := time.Date(2023, 8, 25, 12, 0, 0, 0, time.UTC)
	clock := NewTestClock(testTime)
	repo := New(mockClient, clock)

	userID := int64(1)
	historyRecords := []history.History{
		{ID: int64(1), UserID: userID, Segment: "segment1", Operation: "Added", Time: testTime},
		{ID: int64(2), UserID: userID, Segment: "segment1", Operation: "Deleted", Time: testTime},
	}

	type args struct {
		userID int64
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
					ExpectQuery("SELECT history_id, user_id, segment_name, operation, operation_timestamp FROM segment_history JOIN segments").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			args:     args{userID: userID},
			isError:  false,
			expected: historyRecords,
		},
		{
			title: "Database internal error",
			mockCall: func() {
				mockClient.
					ExpectQuery("SELECT history_id, user_id, segment_name, operation, operation_timestamp FROM segment_history JOIN segments").
					WithArgs(userID).
					WillReturnError(errors.New("internal database error"))
			},
			args:        args{userID: userID},
			isError:     true,
			expected:    nil,
			expectedErr: errors.New("internal database error"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			result, err := repo.GetUserSegments(ctx, test.args.userID)
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
