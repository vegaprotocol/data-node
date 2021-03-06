// Code generated by MockGen. DO NOT EDIT.
// Source: code.vegaprotocol.io/data-node/sqlsubscribers (interfaces: StakeLinkingStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	entities "code.vegaprotocol.io/data-node/entities"
	gomock "github.com/golang/mock/gomock"
)

// MockStakeLinkingStore is a mock of StakeLinkingStore interface.
type MockStakeLinkingStore struct {
	ctrl     *gomock.Controller
	recorder *MockStakeLinkingStoreMockRecorder
}

// MockStakeLinkingStoreMockRecorder is the mock recorder for MockStakeLinkingStore.
type MockStakeLinkingStoreMockRecorder struct {
	mock *MockStakeLinkingStore
}

// NewMockStakeLinkingStore creates a new mock instance.
func NewMockStakeLinkingStore(ctrl *gomock.Controller) *MockStakeLinkingStore {
	mock := &MockStakeLinkingStore{ctrl: ctrl}
	mock.recorder = &MockStakeLinkingStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStakeLinkingStore) EXPECT() *MockStakeLinkingStoreMockRecorder {
	return m.recorder
}

// Upsert mocks base method.
func (m *MockStakeLinkingStore) Upsert(arg0 context.Context, arg1 *entities.StakeLinking) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upsert", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upsert indicates an expected call of Upsert.
func (mr *MockStakeLinkingStoreMockRecorder) Upsert(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upsert", reflect.TypeOf((*MockStakeLinkingStore)(nil).Upsert), arg0, arg1)
}
