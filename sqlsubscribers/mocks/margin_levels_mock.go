// Code generated by MockGen. DO NOT EDIT.
// Source: code.vegaprotocol.io/data-node/sqlsubscribers (interfaces: MarginLevelsStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	entities "code.vegaprotocol.io/data-node/entities"
	gomock "github.com/golang/mock/gomock"
)

// MockMarginLevelsStore is a mock of MarginLevelsStore interface.
type MockMarginLevelsStore struct {
	ctrl     *gomock.Controller
	recorder *MockMarginLevelsStoreMockRecorder
}

// MockMarginLevelsStoreMockRecorder is the mock recorder for MockMarginLevelsStore.
type MockMarginLevelsStoreMockRecorder struct {
	mock *MockMarginLevelsStore
}

// NewMockMarginLevelsStore creates a new mock instance.
func NewMockMarginLevelsStore(ctrl *gomock.Controller) *MockMarginLevelsStore {
	mock := &MockMarginLevelsStore{ctrl: ctrl}
	mock.recorder = &MockMarginLevelsStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMarginLevelsStore) EXPECT() *MockMarginLevelsStoreMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockMarginLevelsStore) Add(arg0 entities.MarginLevels) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockMarginLevelsStoreMockRecorder) Add(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockMarginLevelsStore)(nil).Add), arg0)
}

// Flush mocks base method.
func (m *MockMarginLevelsStore) Flush(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Flush", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Flush indicates an expected call of Flush.
func (mr *MockMarginLevelsStoreMockRecorder) Flush(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Flush", reflect.TypeOf((*MockMarginLevelsStore)(nil).Flush), arg0)
}
