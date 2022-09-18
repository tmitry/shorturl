// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/tmitry/shorturl/internal/app/repositories (interfaces: Repository)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
	models "github.com/tmitry/shorturl/internal/app/models"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// FindAllByUserID mocks base method.
func (m *MockRepository) FindAllByUserID(arg0 uuid.UUID) []*models.ShortURL {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindAllByUserID", arg0)
	ret0, _ := ret[0].([]*models.ShortURL)
	return ret0
}

// FindAllByUserID indicates an expected call of FindAllByUserID.
func (mr *MockRepositoryMockRecorder) FindAllByUserID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindAllByUserID", reflect.TypeOf((*MockRepository)(nil).FindAllByUserID), arg0)
}

// FindOneByUID mocks base method.
func (m *MockRepository) FindOneByUID(arg0 models.UID) *models.ShortURL {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindOneByUID", arg0)
	ret0, _ := ret[0].(*models.ShortURL)
	return ret0
}

// FindOneByUID indicates an expected call of FindOneByUID.
func (mr *MockRepositoryMockRecorder) FindOneByUID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOneByUID", reflect.TypeOf((*MockRepository)(nil).FindOneByUID), arg0)
}

// ReserveID mocks base method.
func (m *MockRepository) ReserveID() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReserveID")
	ret0, _ := ret[0].(int)
	return ret0
}

// ReserveID indicates an expected call of ReserveID.
func (mr *MockRepositoryMockRecorder) ReserveID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReserveID", reflect.TypeOf((*MockRepository)(nil).ReserveID))
}

// Save mocks base method.
func (m *MockRepository) Save(arg0 *models.ShortURL) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Save", arg0)
}

// Save indicates an expected call of Save.
func (mr *MockRepositoryMockRecorder) Save(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockRepository)(nil).Save), arg0)
}