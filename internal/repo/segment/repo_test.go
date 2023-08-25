package segment

import (
	"context"
	"errors"
	"testing"

	"github.com/VrMolodyakov/segment-api/internal/domain/segment/model"
	"github.com/VrMolodyakov/segment-api/internal/domain/segment/service"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
)

func TestCreateSegment(t *testing.T) {
	ctx := context.Background()
	mockPSQLClient, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockPSQLClient.Close()
	repo := New(mockPSQLClient)

	type args struct {
		segment model.Segment
	}

	segmentID := int64(1)

	newSegment := model.Segment{
		Name: "discount",
	}

	tests := []struct {
		title         string
		args          args
		isError       bool
		expected      int64
		expectedError error
		mockCall      func()
	}{
		{
			title: "Should successfully insert a new segment",
			args: args{
				segment: newSegment,
			},
			isError: false,
			mockCall: func() {
				rows := pgxmock.NewRows([]string{"segment_id"}).AddRow(segmentID)
				mockPSQLClient.
					ExpectQuery("INSERT INTO segments").
					WithArgs(newSegment.Name).
					WillReturnRows(rows)
			},
			expected: segmentID,
		},
		{
			title: "Database internal error",
			args: args{
				segment: newSegment,
			},
			isError: true,
			mockCall: func() {
				mockPSQLClient.
					ExpectQuery("INSERT INTO segments").
					WithArgs(newSegment.Name).
					WillReturnError(errors.New("internal database error"))
			},
			expected:      int64(0),
			expectedError: errors.New("internal database error"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			got, err := repo.Create(ctx, test.args.segment)
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

func TestDeleteSegment(t *testing.T) {
	ctx := context.Background()
	mockPSQLClient, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockPSQLClient.Close()
	repo := New(mockPSQLClient)

	segmentName := "discount"

	tests := []struct {
		title         string
		name          string
		isError       bool
		expectedError error
		mockCall      func()
	}{
		{
			title: "Should successfully delete an existing segment",
			name:  segmentName,
			mockCall: func() {
				result := pgxmock.NewResult("DELETE", 1)
				mockPSQLClient.
					ExpectExec("DELETE FROM segments").
					WithArgs(segmentName).
					WillReturnResult(result)
			},
			isError: false,
		},
		{
			title: "Segment not found",
			name:  segmentName,
			mockCall: func() {
				result := pgxmock.NewResult("DELETE", 0)
				mockPSQLClient.
					ExpectExec("DELETE FROM segments").
					WithArgs(segmentName).
					WillReturnResult(result)
			},
			isError:       true,
			expectedError: service.ErrSegmentNotFound,
		},
		{
			title: "Database internal error",
			name:  segmentName,
			mockCall: func() {
				mockPSQLClient.
					ExpectExec("DELETE FROM segments").
					WithArgs(segmentName).
					WillReturnError(errors.New("internal database error"))
			},
			isError:       true,
			expectedError: errors.New("internal database error"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			err := repo.Delete(ctx, test.name)
			if test.isError {
				assert.Error(t, err)
				assert.Equal(t, test.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetAllSegments(t *testing.T) {
	ctx := context.Background()
	mockPSQLClient, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockPSQLClient.Close()
	repo := New(mockPSQLClient)

	segments := []model.Segment{
		{ID: int64(1), Name: "segment1"},
		{ID: int64(2), Name: "segment2"},
	}

	tests := []struct {
		title       string
		isError     bool
		expected    []model.Segment
		expectedErr error
		mockCall    func()
	}{
		{
			title: "Should successfully retrieve all segments",
			mockCall: func() {
				rows := pgxmock.NewRows([]string{"segment_id", "segment_name"}).
					AddRow(segments[0].ID, segments[0].Name).
					AddRow(segments[1].ID, segments[1].Name)
				mockPSQLClient.
					ExpectQuery("SELECT segment_id, segment_name FROM segments").
					WillReturnRows(rows)
			},
			isError:  false,
			expected: segments,
		},
		{
			title: "Database internal error",
			mockCall: func() {
				mockPSQLClient.
					ExpectQuery("SELECT segment_id, segment_name FROM segments").
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
			result, err := repo.GetAll(ctx)
			if test.isError {
				assert.Error(t, err)
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestGetSegment(t *testing.T) {
	ctx := context.Background()
	mockPSQLClient, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mockPSQLClient.Close()
	repo := New(mockPSQLClient)

	segmentName := "discount"
	segmentID := int64(1)

	expectedSegment := model.Segment{
		ID:   segmentID,
		Name: segmentName,
	}

	tests := []struct {
		title         string
		name          string
		isError       bool
		expected      model.Segment
		expectedError error
		mockCall      func()
	}{
		{
			title:   "Should retrieve an existing segment",
			name:    segmentName,
			isError: false,
			mockCall: func() {
				rows := pgxmock.NewRows([]string{"segment_id", "segment_name"}).
					AddRow(segmentID, segmentName)
				mockPSQLClient.
					ExpectQuery("SELECT segment_id, segment_name FROM segments").
					WithArgs(segmentName).
					WillReturnRows(rows)
			},
			expected: expectedSegment,
		},
		{
			title:   "Segment not found",
			name:    "non_existent_segment",
			isError: true,
			mockCall: func() {
				rows := pgxmock.NewRows([]string{"segment_id", "segment_name"})
				mockPSQLClient.
					ExpectQuery("SELECT segment_id, segment_name FROM segments").
					WithArgs("non_existent_segment").
					WillReturnRows(rows)
			},
			expectedError: service.ErrSegmentNotFound,
		},
		{
			title:   "Database internal error",
			name:    segmentName,
			isError: true,
			mockCall: func() {
				mockPSQLClient.
					ExpectQuery("SELECT segment_id, segment_name FROM segments").
					WithArgs(segmentName).
					WillReturnError(errors.New("internal database error"))
			},
			expectedError: errors.New("internal database error"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			got, err := repo.Get(ctx, test.name)
			if test.isError {
				assert.Error(t, err)
				assert.Equal(t, test.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, got)
			}
		})
	}
}
