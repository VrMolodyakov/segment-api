package segment

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/api_server/segment/mocks"
	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/validator"
	segmentService "github.com/VrMolodyakov/segment-api/internal/domain/segment"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateSegment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockSegmentService(ctrl)
	handler := New(mockService)

	type args struct {
		req CreateSegmentRequest
	}

	segmentReq := CreateSegmentRequest{
		Name:          "segment1",
		HitPercentage: 10,
	}

	invalidSegment := CreateSegmentRequest{
		Name:          "s1",
		HitPercentage: 10,
	}

	newSegmentID := int64(1)
	emptyID := int64(0)

	validateError := validator.ValidateError{
		Field: "Name",
		Tag:   "min",
		Param: "6",
	}

	segmentResponse := CreateSegmentResponse{
		ID:         newSegmentID,
		Name:       "segment1",
		Percentage: 10,
	}

	tests := []struct {
		title            string
		args             args
		exoectedCode     int
		mockCall         func()
		expectedResponse func() string
	}{
		{
			title: "Should successfully create new segment",
			mockCall: func() {
				mockService.
					EXPECT().
					CreateSegment(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(newSegmentID, nil)
			},
			expectedResponse: func() string {
				expectedJSON, err := json.Marshal(segmentResponse)
				assert.NoError(t, err)
				return string(expectedJSON)
			},
			args: args{
				req: segmentReq,
			},
			exoectedCode: 201,
		},
		{
			title: "Validate segment error",
			mockCall: func() {
			},
			args: args{
				req: invalidSegment,
			},
			expectedResponse: func() string {
				expectedJSON, err := json.Marshal([]validator.ValidateError{validateError})
				assert.NoError(t, err)
				return string(expectedJSON) + "\n"
			},
			exoectedCode: 400,
		},
		{
			title: "Segment already exists",
			mockCall: func() {
				mockService.
					EXPECT().
					CreateSegment(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(emptyID, segmentService.ErrSegmentAlreadyExists)
			},
			args: args{
				req: segmentReq,
			},
			expectedResponse: func() string {
				return "Segment already exists\n"
			},
			exoectedCode: 400,
		},
		{
			title: "Service error",
			mockCall: func() {
				mockService.
					EXPECT().
					CreateSegment(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(emptyID, errors.New("service error"))
			},
			args: args{
				req: segmentReq,
			},
			expectedResponse: func() string {
				return "Create segment error\n"
			},
			exoectedCode: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			w := httptest.NewRecorder()
			reqBody, err := json.Marshal(test.args.req)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)
			handler.CreateSegment(w, req)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedResponse(), w.Body.String())
			assert.Equal(t, test.exoectedCode, w.Code)

		})
	}
}
