// Code generated by MockGen. DO NOT EDIT.
// Source: code.vegaprotocol.io/data-node/sqlsubscribers (interfaces: MarketsStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	entities "code.vegaprotocol.io/data-node/entities"
	gomock "github.com/golang/mock/gomock"
)

// MockMarketsStore is a mock of MarketsStore interface.
type MockMarketsStore struct {
	ctrl     *gomock.Controller
	recorder *MockMarketsStoreMockRecorder
}

// MockMarketsStoreMockRecorder is the mock recorder for MockMarketsStore.
type MockMarketsStoreMockRecorder struct {
	mock *MockMarketsStore
}

// NewMockMarketsStore creates a new mock instance.
func NewMockMarketsStore(ctrl *gomock.Controller) *MockMarketsStore {
	mock := &MockMarketsStore{ctrl: ctrl}
	mock.recorder = &MockMarketsStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMarketsStore) EXPECT() *MockMarketsStoreMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockMarketsStore) Add(arg0 *entities.Market) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockMarketsStoreMockRecorder) Add(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockMarketsStore)(nil).Add), arg0)
}

// Update mocks base method.
func (m *MockMarketsStore) Update(arg0 *entities.Market) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockMarketsStoreMockRecorder) Update(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockMarketsStore)(nil).Update), arg0)
}
