// Code generated by MockGen. DO NOT EDIT.
// Source: code.vegaprotocol.io/data-node/sqlsubscribers (interfaces: KeyRotationStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	entities "code.vegaprotocol.io/data-node/entities"
	gomock "github.com/golang/mock/gomock"
)

// MockKeyRotationStore is a mock of KeyRotationStore interface.
type MockKeyRotationStore struct {
	ctrl     *gomock.Controller
	recorder *MockKeyRotationStoreMockRecorder
}

// MockKeyRotationStoreMockRecorder is the mock recorder for MockKeyRotationStore.
type MockKeyRotationStoreMockRecorder struct {
	mock *MockKeyRotationStore
}

// NewMockKeyRotationStore creates a new mock instance.
func NewMockKeyRotationStore(ctrl *gomock.Controller) *MockKeyRotationStore {
	mock := &MockKeyRotationStore{ctrl: ctrl}
	mock.recorder = &MockKeyRotationStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKeyRotationStore) EXPECT() *MockKeyRotationStoreMockRecorder {
	return m.recorder
}

// Upsert mocks base method.
func (m *MockKeyRotationStore) Upsert(arg0 context.Context, arg1 *entities.KeyRotation) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upsert", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upsert indicates an expected call of Upsert.
func (mr *MockKeyRotationStoreMockRecorder) Upsert(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upsert", reflect.TypeOf((*MockKeyRotationStore)(nil).Upsert), arg0, arg1)
}
