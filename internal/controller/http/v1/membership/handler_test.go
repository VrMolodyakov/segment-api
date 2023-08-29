package membership

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/membership/mocks"
	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/validator"
	"github.com/VrMolodyakov/segment-api/internal/domain/membership"
	"github.com/VrMolodyakov/segment-api/internal/domain/segment"
	"github.com/VrMolodyakov/segment-api/internal/domain/user"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func AddChiURLParams(r *http.Request, params map[string]string) *http.Request {
	ctx := chi.NewRouteContext()
	for k, v := range params {
		ctx.URLParams.Add(k, v)
	}

	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
}

func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockMembershipService(ctrl)
	handler := New(mockService)

	userID := int64(1)
	emptyID := int64(0)

	type args struct {
		req CreateUserRequest
	}
	newUser := user.User{FirstName: "Bob", LastName: "Bob", Email: "email@email.com"}
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
				mockService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(userID, nil)
			},
			expectedResponse: func() string {
				userResponse := NewCreateUserResponse(userID, newUser.FirstName, newUser.LastName, newUser.Email)
				expectedJSON, err := json.Marshal(userResponse)
				assert.NoError(t, err)
				return string(expectedJSON)
			},
			args: args{
				CreateUserRequest{FirstName: "Bob", LastName: "Bob", Email: "email@email.com"},
			},
			exoectedCode: 201,
		},
		{
			title: "Validate user error",
			mockCall: func() {
			},
			expectedResponse: func() string {
				expectedJSON, err := json.Marshal([]validator.ValidateError{
					{
						Field: "FirstName",
						Tag:   "min",
						Param: "3",
					},
					{
						Field: "LastName",
						Tag:   "min",
						Param: "3",
					},
					{
						Field: "Email",
						Tag:   "min",
						Param: "5",
					},
				})
				assert.NoError(t, err)
				return string(expectedJSON) + "\n"
			},
			args: args{
				CreateUserRequest{FirstName: "a", LastName: "b", Email: "c"},
			},
			exoectedCode: 400,
		},
		{
			title: "Envalid email error",
			mockCall: func() {
				mockService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(emptyID, user.ErrInvalidEmail)
			},
			expectedResponse: func() string {
				return "Invalid email: email validation error.include at least 1 symbol before @ and 2 symbols after\n"
			},
			args: args{
				CreateUserRequest{FirstName: "Bob", LastName: "Bob", Email: "email@email.com"},
			},
			exoectedCode: 400,
		},
		{
			title: "user already exists",
			mockCall: func() {
				mockService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(emptyID, user.ErrUserAlreadyExist)
			},
			expectedResponse: func() string {
				return "User already exists\n"
			},
			args: args{
				CreateUserRequest{FirstName: "Bob", LastName: "Bob", Email: "email@email.com"},
			},
			exoectedCode: 400,
		},
		{
			title: "Error while creating new user",
			mockCall: func() {
				mockService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(emptyID, errors.New("service error"))
			},
			expectedResponse: func() string {
				return "Create user error\n"
			},
			args: args{
				CreateUserRequest{FirstName: "Bob", LastName: "Bob", Email: "email@email.com"},
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
			handler.CreateUser(w, req)
			assert.Equal(t, test.expectedResponse(), w.Body.String())
			assert.Equal(t, test.exoectedCode, w.Code)

		})
	}
}

func TestUpdateUserSegments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockMembershipService(ctrl)
	handler := New(mockService)

	userID := int64(1)

	type args struct {
		req UpdateUserRequest
	}

	tests := []struct {
		title            string
		args             args
		exoectedCode     int
		mockCall         func()
		expectedResponse func() string
	}{
		{
			title: "Should successfully update user segments",
			mockCall: func() {
				mockService.EXPECT().UpdateUserMembership(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedResponse: func() string {
				return ""
			},
			args: args{
				UpdateUserRequest{
					UserID: userID,
					Update: []UpdateSegment{{"segment-1", 10}, {"segment-2", 20}},
					Delete: []string{"segment-3", "segment-4"},
				},
			},
			exoectedCode: 200,
		},
		{
			title: "Validate request error",
			mockCall: func() {
			},
			expectedResponse: func() string {
				expectedJSON, err := json.Marshal([]validator.ValidateError{
					{
						Field: "UserID",
						Tag:   "gt",
						Param: "0",
					},
					{
						Field: "Name",
						Tag:   "min",
						Param: "6",
					},
				})
				assert.NoError(t, err)
				return string(expectedJSON) + "\n"
			},

			args: args{
				UpdateUserRequest{
					UserID: int64(-1),
					Update: []UpdateSegment{{"s", 10}, {"segment-2", 20}},
					Delete: []string{"segment-3", "segment-4"},
				},
			},
			exoectedCode: 400,
		},
		{
			title: "Segment already assigned to user",
			mockCall: func() {
				mockService.EXPECT().UpdateUserMembership(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(membership.ErrSegmentAlreadyAssigned)
			},
			expectedResponse: func() string {
				return "Attempt to add segments that the user already belongs to\n"
			},
			args: args{
				UpdateUserRequest{
					UserID: userID,
					Update: []UpdateSegment{{"segment-1", 10}, {"segment-2", 20}},
					Delete: []string{"segment-3", "segment-4"},
				},
			},
			exoectedCode: 400,
		},
		{
			title: "Segment does not exists",
			mockCall: func() {
				mockService.EXPECT().UpdateUserMembership(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(segment.ErrSegmentNotFound)
			},
			expectedResponse: func() string {
				return "Not all segments with the specified names were found\n"
			},
			args: args{
				UpdateUserRequest{
					UserID: userID,
					Update: []UpdateSegment{{"segment-1", 10}, {"segment-2", 20}},
					Delete: []string{"segment-3", "segment-4"},
				},
			},
			exoectedCode: 400,
		},
		{
			title: "User does not exists",
			mockCall: func() {
				mockService.EXPECT().UpdateUserMembership(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(user.ErrUserNotFound)
			},
			expectedResponse: func() string {
				return "Attempt to update the data of a non-existent user\n"
			},
			args: args{
				UpdateUserRequest{
					UserID: userID,
					Update: []UpdateSegment{{"segment-1", 10}, {"segment-2", 20}},
					Delete: []string{"segment-3", "segment-4"},
				},
			},
			exoectedCode: 400,
		},
		{
			title: "Both add and delete array is empty",
			mockCall: func() {
				mockService.EXPECT().UpdateUserMembership(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(membership.ErrEmptyData)
			},
			expectedResponse: func() string {
				return "Data for update and delete cannot be empty at the same time\n"
			},
			args: args{
				UpdateUserRequest{
					UserID: userID,
					Update: []UpdateSegment{},
					Delete: []string{},
				},
			},
			exoectedCode: 400,
		},
		{
			title: "Attempty to add and delete the same segment",
			mockCall: func() {
				mockService.EXPECT().UpdateUserMembership(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(membership.ErrIncorrectData)
			},
			expectedResponse: func() string {
				return "Attempt to add and remove the same segment\n"
			},
			args: args{
				UpdateUserRequest{
					UserID: userID,
					Update: []UpdateSegment{{"segment-1", 10}, {"segment-2", 20}},
					Delete: []string{"segment-3", "segment-4"},
				},
			},
			exoectedCode: 400,
		},
		{
			title: "Error while updating user segments",
			mockCall: func() {
				mockService.EXPECT().UpdateUserMembership(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("service error"))
			},
			expectedResponse: func() string {
				return "Update user segments\n"
			},
			args: args{
				UpdateUserRequest{
					UserID: userID,
					Update: []UpdateSegment{{"segment-4", 10}, {"segment-2", 20}},
					Delete: []string{"segment-3", "segment-4"},
				},
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
			handler.UpdateUserMembership(w, req)
			assert.Equal(t, test.exoectedCode, w.Code)
			assert.Equal(t, test.expectedResponse(), w.Body.String())
		})
	}
}

func TestDeleteMembership(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockMembershipService(ctrl)
	handler := New(mockService)

	type args struct {
		req DeleteSegmentRequest
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
				mockService.EXPECT().DeleteMembership(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedResponse: func() string {
				return ""
			},
			args: args{
				DeleteSegmentRequest{"segment-1"},
			},
			exoectedCode: 200,
		},
		{
			title: "Validate user error",
			mockCall: func() {
			},
			expectedResponse: func() string {
				expectedJSON, err := json.Marshal([]validator.ValidateError{
					{
						Field: "Name",
						Tag:   "min",
						Param: "6",
					},
				})
				assert.NoError(t, err)
				return string(expectedJSON) + "\n"
			},
			args: args{
				DeleteSegmentRequest{"s"},
			},
			exoectedCode: 400,
		},
		{
			title: "Segment not found",
			mockCall: func() {
				mockService.EXPECT().DeleteMembership(gomock.Any(), gomock.Any()).Return(segment.ErrSegmentNotFound)
			},
			expectedResponse: func() string {
				return "Segment with the specified name wasn't found\n"
			},
			args: args{
				DeleteSegmentRequest{"segment-1"},
			},
			exoectedCode: 400,
		},
		{
			title: "Service error",
			mockCall: func() {
				mockService.EXPECT().DeleteMembership(gomock.Any(), gomock.Any()).Return(errors.New("service error"))
			},
			expectedResponse: func() string {
				return "Delete segment\n"
			},
			args: args{
				DeleteSegmentRequest{"segment-1"},
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
			handler.DeleteMembership(w, req)
			assert.Equal(t, test.expectedResponse(), w.Body.String())
			assert.Equal(t, test.exoectedCode, w.Code)

		})
	}
}

func TestGetUserMembership(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockMembershipService(ctrl)
	handler := New(mockService)

	type args struct {
		param map[string]string
	}

	info := []membership.MembershipInfo{{UserID: 1, SegmentName: "seg-1"}, {UserID: 2, SegmentName: "seg-2"}}

	tests := []struct {
		title            string
		exoectedCode     int
		args             args
		mockCall         func()
		expectedResponse func() string
	}{
		{
			title: "Should successfully create new segment",
			mockCall: func() {
				mockService.EXPECT().GetUserMembership(gomock.Any(), gomock.Any()).Return(info, nil)
			},
			args: args{
				param: map[string]string{
					"userID": "1",
				},
			},
			expectedResponse: func() string {
				response := GetUserMembershipResponse{
					Memberships: []UserResponseInfo{{UserID: 1, SegmentName: "seg-1"}, {UserID: 2, SegmentName: "seg-2"}},
				}
				expectedJSON, err := json.Marshal(response)
				assert.NoError(t, err)
				return string(expectedJSON)

			},
			exoectedCode: 200,
		},
		{
			title: "Incorrect user id error",
			mockCall: func() {
			},
			args: args{
				param: map[string]string{
					"userID": "abc",
				},
			},
			expectedResponse: func() string {
				return "Invalid year parameter\n"

			},
			exoectedCode: 400,
		},
		{
			title: "Service error",
			mockCall: func() {
				mockService.EXPECT().GetUserMembership(gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))
			},
			args: args{
				param: map[string]string{
					"userID": "1",
				},
			},
			expectedResponse: func() string {
				return "Get user membership segment\n"

			},
			exoectedCode: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			w := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, "", nil)
			assert.NoError(t, err)
			req = AddChiURLParams(req, test.args.param)

			handler.GetUserMembership(w, req)
			assert.Equal(t, test.expectedResponse(), w.Body.String())
			assert.Equal(t, test.exoectedCode, w.Code)

		})
	}
}
