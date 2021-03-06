// Code generated by MockGen. DO NOT EDIT.
// Source: code.vegaprotocol.io/data-node/sqlsubscribers (interfaces: OracleSpecStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	entities "code.vegaprotocol.io/data-node/entities"
	gomock "github.com/golang/mock/gomock"
)

// MockOracleSpecStore is a mock of OracleSpecStore interface.
type MockOracleSpecStore struct {
	ctrl     *gomock.Controller
	recorder *MockOracleSpecStoreMockRecorder
}

// MockOracleSpecStoreMockRecorder is the mock recorder for MockOracleSpecStore.
type MockOracleSpecStoreMockRecorder struct {
	mock *MockOracleSpecStore
}

// NewMockOracleSpecStore creates a new mock instance.
func NewMockOracleSpecStore(ctrl *gomock.Controller) *MockOracleSpecStore {
	mock := &MockOracleSpecStore{ctrl: ctrl}
	mock.recorder = &MockOracleSpecStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOracleSpecStore) EXPECT() *MockOracleSpecStoreMockRecorder {
	return m.recorder
}

// Upsert mocks base method.
func (m *MockOracleSpecStore) Upsert(arg0 context.Context, arg1 *entities.OracleSpec) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upsert", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upsert indicates an expected call of Upsert.
func (mr *MockOracleSpecStoreMockRecorder) Upsert(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upsert", reflect.TypeOf((*MockOracleSpecStore)(nil).Upsert), arg0, arg1)
}
