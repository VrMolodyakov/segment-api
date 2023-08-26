// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/domain/participation/service/service.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	model "github.com/VrMolodyakov/segment-api/internal/domain/participation/model"
	model0 "github.com/VrMolodyakov/segment-api/internal/domain/segment/model"
	gomock "github.com/golang/mock/gomock"
)

// MockParticipationRepository is a mock of ParticipationRepository interface.
type MockParticipationRepository struct {
	ctrl     *gomock.Controller
	recorder *MockParticipationRepositoryMockRecorder
}

// MockParticipationRepositoryMockRecorder is the mock recorder for MockParticipationRepository.
type MockParticipationRepositoryMockRecorder struct {
	mock *MockParticipationRepository
}

// NewMockParticipationRepository creates a new mock instance.
func NewMockParticipationRepository(ctrl *gomock.Controller) *MockParticipationRepository {
	mock := &MockParticipationRepository{ctrl: ctrl}
	mock.recorder = &MockParticipationRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockParticipationRepository) EXPECT() *MockParticipationRepositoryMockRecorder {
	return m.recorder
}

// DeleteSegment mocks base method.
func (m *MockParticipationRepository) DeleteSegment(ctx context.Context, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSegment", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSegment indicates an expected call of DeleteSegment.
func (mr *MockParticipationRepositoryMockRecorder) DeleteSegment(ctx, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSegment", reflect.TypeOf((*MockParticipationRepository)(nil).DeleteSegment), ctx, name)
}

// GetUserSegments mocks base method.
func (m *MockParticipationRepository) GetUserSegments(ctx context.Context, userID int64) ([]model.Participation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserSegments", ctx, userID)
	ret0, _ := ret[0].([]model.Participation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserSegments indicates an expected call of GetUserSegments.
func (mr *MockParticipationRepositoryMockRecorder) GetUserSegments(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserSegments", reflect.TypeOf((*MockParticipationRepository)(nil).GetUserSegments), ctx, userID)
}

// UpdateUserSegments mocks base method.
func (m *MockParticipationRepository) UpdateUserSegments(ctx context.Context, userID int64, addSegments []model0.Segment, deleteSegments []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUserSegments", ctx, userID, addSegments, deleteSegments)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUserSegments indicates an expected call of UpdateUserSegments.
func (mr *MockParticipationRepositoryMockRecorder) UpdateUserSegments(ctx, userID, addSegments, deleteSegments interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUserSegments", reflect.TypeOf((*MockParticipationRepository)(nil).UpdateUserSegments), ctx, userID, addSegments, deleteSegments)
}
