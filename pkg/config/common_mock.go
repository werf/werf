// Code generated by MockGen. DO NOT EDIT.
// Source: common.go
//
// Generated by this command:
//
//	mockgen -source common.go -package config -destination common_mock.go
//

// Package config is a generated GoMock package.
package config

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockrawOrigin is a mock of rawOrigin interface.
type MockrawOrigin struct {
	ctrl     *gomock.Controller
	recorder *MockrawOriginMockRecorder
	isgomock struct{}
}

// MockrawOriginMockRecorder is the mock recorder for MockrawOrigin.
type MockrawOriginMockRecorder struct {
	mock *MockrawOrigin
}

// NewMockrawOrigin creates a new mock instance.
func NewMockrawOrigin(ctrl *gomock.Controller) *MockrawOrigin {
	mock := &MockrawOrigin{ctrl: ctrl}
	mock.recorder = &MockrawOriginMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockrawOrigin) EXPECT() *MockrawOriginMockRecorder {
	return m.recorder
}

// configSection mocks base method.
func (m *MockrawOrigin) configSection() any {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "configSection")
	ret0, _ := ret[0].(any)
	return ret0
}

// configSection indicates an expected call of configSection.
func (mr *MockrawOriginMockRecorder) configSection() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "configSection", reflect.TypeOf((*MockrawOrigin)(nil).configSection))
}

// doc mocks base method.
func (m *MockrawOrigin) doc() *doc {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "doc")
	ret0, _ := ret[0].(*doc)
	return ret0
}

// doc indicates an expected call of doc.
func (mr *MockrawOriginMockRecorder) doc() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "doc", reflect.TypeOf((*MockrawOrigin)(nil).doc))
}
