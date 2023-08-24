package segment

import (
	"context"
	"errors"
	"testing"

	"github.com/VrMolodyakov/segment-api/internal/domain/segment/model"
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

	segmentID := 1

	newSegment := model.Segment{
		Name: "discount",
	}

	tests := []struct {
		title         string
		args          args
		isError       bool
		expected      int
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
			expected:      0,
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
