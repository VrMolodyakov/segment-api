// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/controller/http/v1/membership/handler.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	membership "github.com/VrMolodyakov/segment-api/internal/domain/membership"
	segment "github.com/VrMolodyakov/segment-api/internal/domain/segment"
	user "github.com/VrMolodyakov/segment-api/internal/domain/user"
	gomock "github.com/golang/mock/gomock"
)

// MockMembershipService is a mock of MembershipService interface.
type MockMembershipService struct {
	ctrl     *gomock.Controller
	recorder *MockMembershipServiceMockRecorder
}

// MockMembershipServiceMockRecorder is the mock recorder for MockMembershipService.
type MockMembershipServiceMockRecorder struct {
	mock *MockMembershipService
}

// NewMockMembershipService creates a new mock instance.
func NewMockMembershipService(ctrl *gomock.Controller) *MockMembershipService {
	mock := &MockMembershipService{ctrl: ctrl}
	mock.recorder = &MockMembershipServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMembershipService) EXPECT() *MockMembershipServiceMockRecorder {
	return m.recorder
}

// CreateUser mocks base method.
func (m *MockMembershipService) CreateUser(ctx context.Context, user user.User) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", ctx, user)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockMembershipServiceMockRecorder) CreateUser(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockMembershipService)(nil).CreateUser), ctx, user)
}

// DeleteMembership mocks base method.
func (m *MockMembershipService) DeleteMembership(ctx context.Context, segmentName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteMembership", ctx, segmentName)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteMembership indicates an expected call of DeleteMembership.
func (mr *MockMembershipServiceMockRecorder) DeleteMembership(ctx, segmentName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMembership", reflect.TypeOf((*MockMembershipService)(nil).DeleteMembership), ctx, segmentName)
}

// GetUserMembership mocks base method.
func (m *MockMembershipService) GetUserMembership(ctx context.Context, userID int64) ([]membership.MembershipInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserMembership", ctx, userID)
	ret0, _ := ret[0].([]membership.MembershipInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserMembership indicates an expected call of GetUserMembership.
func (mr *MockMembershipServiceMockRecorder) GetUserMembership(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserMembership", reflect.TypeOf((*MockMembershipService)(nil).GetUserMembership), ctx, userID)
}

// UpdateUserMembership mocks base method.
func (m *MockMembershipService) UpdateUserMembership(ctx context.Context, userID int64, addSegments []segment.Segment, deleteSegments []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUserMembership", ctx, userID, addSegments, deleteSegments)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUserMembership indicates an expected call of UpdateUserMembership.
func (mr *MockMembershipServiceMockRecorder) UpdateUserMembership(ctx, userID, addSegments, deleteSegments interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUserMembership", reflect.TypeOf((*MockMembershipService)(nil).UpdateUserMembership), ctx, userID, addSegments, deleteSegments)
}
