// Code generated by MockGen. DO NOT EDIT.
// Source: code.vegaprotocol.io/data-node/sqlsubscribers (interfaces: LiquidityProvisionStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	entities "code.vegaprotocol.io/data-node/entities"
	gomock "github.com/golang/mock/gomock"
)

// MockLiquidityProvisionStore is a mock of LiquidityProvisionStore interface.
type MockLiquidityProvisionStore struct {
	ctrl     *gomock.Controller
	recorder *MockLiquidityProvisionStoreMockRecorder
}

// MockLiquidityProvisionStoreMockRecorder is the mock recorder for MockLiquidityProvisionStore.
type MockLiquidityProvisionStoreMockRecorder struct {
	mock *MockLiquidityProvisionStore
}

// NewMockLiquidityProvisionStore creates a new mock instance.
func NewMockLiquidityProvisionStore(ctrl *gomock.Controller) *MockLiquidityProvisionStore {
	mock := &MockLiquidityProvisionStore{ctrl: ctrl}
	mock.recorder = &MockLiquidityProvisionStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLiquidityProvisionStore) EXPECT() *MockLiquidityProvisionStoreMockRecorder {
	return m.recorder
}

// Flush mocks base method.
func (m *MockLiquidityProvisionStore) Flush(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Flush", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Flush indicates an expected call of Flush.
func (mr *MockLiquidityProvisionStoreMockRecorder) Flush(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Flush", reflect.TypeOf((*MockLiquidityProvisionStore)(nil).Flush), arg0)
}

// Upsert mocks base method.
func (m *MockLiquidityProvisionStore) Upsert(arg0 context.Context, arg1 entities.LiquidityProvision) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upsert", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upsert indicates an expected call of Upsert.
func (mr *MockLiquidityProvisionStoreMockRecorder) Upsert(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upsert", reflect.TypeOf((*MockLiquidityProvisionStore)(nil).Upsert), arg0, arg1)
}
