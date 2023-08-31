package membership

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/apierror"
	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver/membership/mocks"
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
				resp, err := json.Marshal(apierror.ErrorResponse{Message: string(expectedJSON)})
				assert.NoError(t, err)
				return string(resp)
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
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "Invalid email: email validation error.include at least 1 symbol before @ and 2 symbols after and dot (example@example.com)"})
				assert.NoError(t, err)
				return string(resp)
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
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "User already exists"})
				assert.NoError(t, err)
				return string(resp)
			},
			args: args{
				CreateUserRequest{FirstName: "Bob", LastName: "Bob", Email: "email@email.com"},
			},
			exoectedCode: 409,
		},
		{
			title: "Error while creating new user",
			mockCall: func() {
				mockService.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(emptyID, errors.New("service error"))
			},
			expectedResponse: func() string {
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "Create user error"})
				assert.NoError(t, err)
				return string(resp)
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
					Delete: []DeleteSegment{{"segment-3"}, {"segment-4"}},
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
				resp, err := json.Marshal(apierror.ErrorResponse{Message: string(expectedJSON)})
				assert.NoError(t, err)
				return string(resp)
			},

			args: args{
				UpdateUserRequest{
					UserID: int64(-1),
					Update: []UpdateSegment{{"s", 10}, {"segment-2", 20}},
					Delete: []DeleteSegment{{"segment-3"}, {"segment-4"}},
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
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "Attempt to add segments that the user already belongs to"})
				assert.NoError(t, err)
				return string(resp)
			},
			args: args{
				UpdateUserRequest{
					UserID: userID,
					Update: []UpdateSegment{{"segment-1", 10}, {"segment-2", 20}},
					Delete: []DeleteSegment{{"segment-3"}, {"segment-4"}},
				},
			},
			exoectedCode: 409,
		},
		{
			title: "Segment does not exists",
			mockCall: func() {
				mockService.EXPECT().UpdateUserMembership(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(segment.ErrSegmentNotFound)
			},
			expectedResponse: func() string {
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "Not all segments with the specified names were found or adding/removing one segment multiple times"})
				assert.NoError(t, err)
				return string(resp)
			},
			args: args{
				UpdateUserRequest{
					UserID: userID,
					Update: []UpdateSegment{{"segment-1", 10}, {"segment-2", 20}},
					Delete: []DeleteSegment{{"segment-3"}, {"segment-4"}},
				},
			},
			exoectedCode: 404,
		},
		{
			title: "User does not exists",
			mockCall: func() {
				mockService.EXPECT().UpdateUserMembership(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(user.ErrUserNotFound)
			},
			expectedResponse: func() string {
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "Attempt to update the data of a non-existent user"})
				assert.NoError(t, err)
				return string(resp)
			},
			args: args{
				UpdateUserRequest{
					UserID: userID,
					Update: []UpdateSegment{{"segment-1", 10}, {"segment-2", 20}},
					Delete: []DeleteSegment{{"segment-3"}, {"segment-4"}},
				},
			},
			exoectedCode: 404,
		},
		{
			title: "Both add and delete array is empty",
			mockCall: func() {
				mockService.EXPECT().UpdateUserMembership(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(membership.ErrEmptyData)
			},
			expectedResponse: func() string {
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "Data for update and delete cannot be empty at the same time"})
				assert.NoError(t, err)
				return string(resp)
			},
			args: args{
				UpdateUserRequest{
					UserID: userID,
					Update: []UpdateSegment{},
					Delete: []DeleteSegment{},
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
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "Attempt to add and remove the same segment"})
				assert.NoError(t, err)
				return string(resp)
			},
			args: args{
				UpdateUserRequest{
					UserID: userID,
					Update: []UpdateSegment{{"segment-1", 10}, {"segment-2", 20}},
					Delete: []DeleteSegment{{"segment-3"}, {"segment-4"}},
				},
			},
			exoectedCode: 400,
		},
		{
			title: "Attempty to delete unassigned segment",
			mockCall: func() {
				mockService.EXPECT().UpdateUserMembership(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(membership.ErrSegmentNotAssigned)
			},
			expectedResponse: func() string {
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "Attempt to delete a segment unassigned to the user"})
				assert.NoError(t, err)
				return string(resp)
			},
			args: args{
				UpdateUserRequest{
					UserID: userID,
					Update: []UpdateSegment{{"segment-1", 10}, {"segment-2", 20}},
					Delete: []DeleteSegment{{"segment-3"}, {"segment-4"}},
				},
			},
			exoectedCode: 409,
		},
		{
			title: "Error while updating user segments",
			mockCall: func() {
				mockService.EXPECT().UpdateUserMembership(
					gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("service error"))
			},
			expectedResponse: func() string {
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "Update user segments"})
				assert.NoError(t, err)
				return string(resp)
			},
			args: args{
				UpdateUserRequest{
					UserID: userID,
					Update: []UpdateSegment{{"segment-4", 10}, {"segment-2", 20}},
					Delete: []DeleteSegment{{"segment-3"}, {"segment-4"}},
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
		param map[string]string
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
				map[string]string{"segmentName": "segment-1"},
			},
			exoectedCode: 200,
		},
		{
			title: "Segment not found",
			mockCall: func() {
				mockService.EXPECT().DeleteMembership(gomock.Any(), gomock.Any()).Return(segment.ErrSegmentNotFound)
			},
			expectedResponse: func() string {
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "Segment with the specified name wasn't found"})
				assert.NoError(t, err)
				return string(resp)
			},
			args: args{
				map[string]string{"segmentName": "segment-1"},
			},
			exoectedCode: 404,
		},
		{
			title: "Service error",
			mockCall: func() {
				mockService.EXPECT().DeleteMembership(gomock.Any(), gomock.Any()).Return(errors.New("service error"))
			},
			expectedResponse: func() string {
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "Delete segment"})
				assert.NoError(t, err)
				return string(resp)
			},
			args: args{
				map[string]string{"segmentName": "segment-1"},
			},
			exoectedCode: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			test.mockCall()
			w := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodDelete, "", nil)
			assert.NoError(t, err)
			req = AddChiURLParams(req, test.args.param)
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
				data := make([]UserResponseInfo, len(info))
				for i := range data {
					data[i] = NewUserResponseInfo(info[i].UserID, info[i].SegmentName, info[i].ExpiredAt)
				}
				expectedJSON, err := json.Marshal(NewUserMembershipResponse(data))
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
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "Invalid user id parameter"})
				assert.NoError(t, err)
				return string(resp)

			},
			exoectedCode: 400,
		},
		{
			title: "Data not found",
			mockCall: func() {
				mockService.EXPECT().GetUserMembership(gomock.Any(), gomock.Any()).Return([]membership.MembershipInfo{}, nil)
			},
			args: args{
				param: map[string]string{
					"userID": "1",
				},
			},
			expectedResponse: func() string {
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "No data was found for the specified user"})
				assert.NoError(t, err)
				return string(resp)

			},
			exoectedCode: 404,
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
				resp, err := json.Marshal(apierror.ErrorResponse{Message: "Get user membership segment"})
				assert.NoError(t, err)
				return string(resp)
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
