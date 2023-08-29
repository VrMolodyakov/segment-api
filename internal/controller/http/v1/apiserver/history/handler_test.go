package history

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/history/mocks"
	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/validator"
	"github.com/VrMolodyakov/segment-api/internal/domain/history"
	"github.com/VrMolodyakov/segment-api/pkg/csv"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type mockBufferPool struct {
}

func (m *mockBufferPool) Get() *bytes.Buffer {
	return &bytes.Buffer{}
}
func (m *mockBufferPool) Release(buf *bytes.Buffer) {}

func AddChiURLParams(r *http.Request, params map[string]string) *http.Request {
	ctx := chi.NewRouteContext()
	for k, v := range params {
		ctx.URLParams.Add(k, v)
	}

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
}

func TestCreateLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockHistoryService(ctrl)

	param := LinkParam{
		Host: "localhost",
		Port: 8080,
	}

	linkRes := CreateLinkResponse{
		Link: "http://localhost:8080/api/v1/history/download/2023/8",
	}

	handler := New(mockService, param, nil, nil)

	type args struct {
		req CreateLinkRequest
	}

	tests := []struct {
		title            string
		args             args
		exoectedCode     int
		mockCall         func()
		expected         CreateLinkResponse
		expectedResponse func() string
	}{
		{
			title: "Should successfully create new segment",
			mockCall: func() {
				mockService.EXPECT().PrepareHistoryData(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedResponse: func() string {
				expectedJSON, err := json.Marshal(linkRes)
				assert.NoError(t, err)
				return string(expectedJSON)
			},
			args: args{
				req: CreateLinkRequest{
					Year:  2023,
					Month: 8,
				},
			},
			expected:     linkRes,
			exoectedCode: 200,
		},
		{
			title: "Invalid year request",
			mockCall: func() {
			},
			expectedResponse: func() string {
				v := validator.ValidateError{
					Field: "Year",
					Tag:   "gt",
					Param: "-1",
				}
				expectedJSON, err := json.Marshal([]validator.ValidateError{v})
				assert.NoError(t, err)
				return string(expectedJSON) + "\n"
			},
			args: args{
				req: CreateLinkRequest{
					Year:  -1,
					Month: 8,
				},
			},
			exoectedCode: 400,
		},
		{
			title: "Invalid month request",
			mockCall: func() {
			},
			expectedResponse: func() string {
				v := validator.ValidateError{
					Field: "Month",
					Tag:   "lt",
					Param: "13",
				}
				expectedJSON, err := json.Marshal([]validator.ValidateError{v})
				assert.NoError(t, err)
				return string(expectedJSON) + "\n"
			},
			args: args{
				req: CreateLinkRequest{
					Year:  2023,
					Month: 13,
				},
			},
			exoectedCode: 400,
		},
		{
			title: "Service incorrect year error",
			mockCall: func() {
				mockService.EXPECT().PrepareHistoryData(gomock.Any(), gomock.Any()).Return(history.ErrIncorrectYear)
			},
			expectedResponse: func() string {
				return "Incorrect date, history for dates before 2007 year is not available\n"
			},
			args: args{
				req: CreateLinkRequest{Year: 1994, Month: 1},
			},
			exoectedCode: 400,
		},
		{
			title: "Service incorrect month error",
			mockCall: func() {
				mockService.EXPECT().PrepareHistoryData(gomock.Any(), gomock.Any()).Return(history.ErrIncorrectMonth)
			},
			expectedResponse: func() string {
				return "Incorrect date, impossible to get information for a month that has not yet come\n"
			},
			args: args{
				req: CreateLinkRequest{Year: 1994, Month: 1},
			},
			exoectedCode: 400,
		},
		{
			title: "Service error",
			mockCall: func() {
				mockService.EXPECT().PrepareHistoryData(gomock.Any(), gomock.Any()).Return(errors.New("service error"))
			},
			expectedResponse: func() string {
				return "Couldn't prepare history data\n"
			},
			args: args{
				req: CreateLinkRequest{Year: 2023, Month: 8},
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
			handler.CreateLink(w, req)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedResponse(), w.Body.String())
			assert.Equal(t, test.exoectedCode, w.Code)

		})
	}
}

func TestDownloadCSVData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockHistoryService(ctrl)

	param := LinkParam{
		Host: "localhost",
		Port: 8080,
	}
	handler := New(mockService, param, &mockBufferPool{}, nil)

	type args struct {
		reqParam map[string]string
	}

	tests := []struct {
		title            string
		args             args
		exoectedCode     int
		writerCallMock   func(w io.Writer, args []history.History) error
		mockCall         func()
		expectedResponse func() string
	}{
		{
			title: "Should successfully create new segment",
			mockCall: func() {
				mockService.EXPECT().GetUsersHistory(gomock.Any(), gomock.Any()).Return([]history.History{}, nil)
			},
			expectedResponse: func() string {
				return "hello segment"
			},
			args: args{
				reqParam: map[string]string{
					"year":  "2023",
					"month": "8",
				},
			},
			writerCallMock: func(w io.Writer, args []history.History) error {
				_, err := w.Write([]byte("hello segment"))
				assert.NoError(t, err)
				return err
			},
			exoectedCode: 200,
		},
		{
			title: "Invalid year URL params",
			mockCall: func() {
			},
			expectedResponse: func() string {
				return "Invalid year parameter\n"
			},
			args: args{
				reqParam: map[string]string{
					"year":  "",
					"month": "8",
				},
			},
			exoectedCode: 400,
		},
		{
			title: "Invalid month URL params",
			mockCall: func() {
			},
			expectedResponse: func() string {
				return "Invalid month parameter\n"
			},
			args: args{
				reqParam: map[string]string{
					"year":  "2023",
					"month": "",
				},
			},
			exoectedCode: 400,
		},
		{
			title: "Service incorrect year error",
			mockCall: func() {
				mockService.EXPECT().GetUsersHistory(gomock.Any(), gomock.Any()).Return(nil, history.ErrIncorrectYear)
			},
			expectedResponse: func() string {
				return "Incorrect date, history for dates before 2007 year is not available\n"
			},
			args: args{
				reqParam: map[string]string{
					"year":  "2023",
					"month": "8",
				},
			},
			exoectedCode: 400,
		},
		{
			title: "Service incorrect month error",
			mockCall: func() {
				mockService.EXPECT().GetUsersHistory(gomock.Any(), gomock.Any()).Return(nil, history.ErrIncorrectMonth)
			},
			expectedResponse: func() string {
				return "Incorrect date, impossible to get information for a month that has not yet come\n"
			},
			args: args{
				reqParam: map[string]string{
					"year":  "2023",
					"month": "8",
				},
			},
			exoectedCode: 400,
		},
		{
			title: "Service data already expired error",
			mockCall: func() {
				mockService.EXPECT().GetUsersHistory(gomock.Any(), gomock.Any()).Return(nil, history.ErrExpiredData)
			},
			expectedResponse: func() string {
				return "Data lifetime for the link has expired, create a new one\n"
			},
			args: args{
				reqParam: map[string]string{
					"year":  "2023",
					"month": "8",
				},
			},
			exoectedCode: 404,
		},
		{
			title: "Couldn't create csv data",
			mockCall: func() {
				mockService.EXPECT().GetUsersHistory(gomock.Any(), gomock.Any()).Return([]history.History{}, nil)
			},
			expectedResponse: func() string {
				return "Couldn't create a csv file, error\n"
			},
			args: args{
				reqParam: map[string]string{
					"year":  "2023",
					"month": "8",
				},
			},
			writerCallMock: func(w io.Writer, args []history.History) error {
				return errors.New("error")
			},
			exoectedCode: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			w := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			assert.NoError(t, err)
			req = AddChiURLParams(req, test.args.reqParam)

			testCSV := csv.NewCSVWriter[history.History](test.writerCallMock)
			handler.writer = &testCSV
			handler.DownloadCSVData(w, req)

			assert.Equal(t, test.expectedResponse(), w.Body.String())
			assert.Equal(t, test.exoectedCode, w.Code)

		})
	}
}
