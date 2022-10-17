// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/tmitry/shorturl/internal/app/utils (interfaces: UIDGenerator)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/tmitry/shorturl/internal/app/models"
)

// MockUIDGenerator is a mock of UIDGenerator interface.
type MockUIDGenerator struct {
	ctrl     *gomock.Controller
	recorder *MockUIDGeneratorMockRecorder
}

// MockUIDGeneratorMockRecorder is the mock recorder for MockUIDGenerator.
type MockUIDGeneratorMockRecorder struct {
	mock *MockUIDGenerator
}

// NewMockUIDGenerator creates a new mock instance.
func NewMockUIDGenerator(ctrl *gomock.Controller) *MockUIDGenerator {
	mock := &MockUIDGenerator{ctrl: ctrl}
	mock.recorder = &MockUIDGeneratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUIDGenerator) EXPECT() *MockUIDGeneratorMockRecorder {
	return m.recorder
}

// Generate mocks base method.
func (m *MockUIDGenerator) Generate() (models.UID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Generate")
	ret0, _ := ret[0].(models.UID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Generate indicates an expected call of Generate.
func (mr *MockUIDGeneratorMockRecorder) Generate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Generate", reflect.TypeOf((*MockUIDGenerator)(nil).Generate))
}

// GetPattern mocks base method.
func (m *MockUIDGenerator) GetPattern() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPattern")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetPattern indicates an expected call of GetPattern.
func (mr *MockUIDGeneratorMockRecorder) GetPattern() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPattern", reflect.TypeOf((*MockUIDGenerator)(nil).GetPattern))
}

// IsValid mocks base method.
func (m *MockUIDGenerator) IsValid(arg0 models.UID) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsValid", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsValid indicates an expected call of IsValid.
func (mr *MockUIDGeneratorMockRecorder) IsValid(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsValid", reflect.TypeOf((*MockUIDGenerator)(nil).IsValid), arg0)
}
