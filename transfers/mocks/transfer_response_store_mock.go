// Code generated by MockGen. DO NOT EDIT.
// Source: code.vegaprotocol.io/data-node/transfers (interfaces: TransferResponseStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	proto "code.vegaprotocol.io/data-node/proto"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockTransferResponseStore is a mock of TransferResponseStore interface
type MockTransferResponseStore struct {
	ctrl     *gomock.Controller
	recorder *MockTransferResponseStoreMockRecorder
}

// MockTransferResponseStoreMockRecorder is the mock recorder for MockTransferResponseStore
type MockTransferResponseStoreMockRecorder struct {
	mock *MockTransferResponseStore
}

// NewMockTransferResponseStore creates a new mock instance
func NewMockTransferResponseStore(ctrl *gomock.Controller) *MockTransferResponseStore {
	mock := &MockTransferResponseStore{ctrl: ctrl}
	mock.recorder = &MockTransferResponseStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTransferResponseStore) EXPECT() *MockTransferResponseStoreMockRecorder {
	return m.recorder
}

// Subscribe mocks base method
func (m *MockTransferResponseStore) Subscribe(arg0 chan []*proto.TransferResponse) uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe", arg0)
	ret0, _ := ret[0].(uint64)
	return ret0
}

// Subscribe indicates an expected call of Subscribe
func (mr *MockTransferResponseStoreMockRecorder) Subscribe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockTransferResponseStore)(nil).Subscribe), arg0)
}

// Unsubscribe mocks base method
func (m *MockTransferResponseStore) Unsubscribe(arg0 uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unsubscribe", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unsubscribe indicates an expected call of Unsubscribe
func (mr *MockTransferResponseStoreMockRecorder) Unsubscribe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unsubscribe", reflect.TypeOf((*MockTransferResponseStore)(nil).Unsubscribe), arg0)
}
