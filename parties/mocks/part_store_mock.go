// Code generated by MockGen. DO NOT EDIT.
// Source: code.vegaprotocol.io/data-node/parties (interfaces: PartyStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	proto "code.vegaprotocol.io/data-node/proto/vega"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockPartyStore is a mock of PartyStore interface
type MockPartyStore struct {
	ctrl     *gomock.Controller
	recorder *MockPartyStoreMockRecorder
}

// MockPartyStoreMockRecorder is the mock recorder for MockPartyStore
type MockPartyStoreMockRecorder struct {
	mock *MockPartyStore
}

// NewMockPartyStore creates a new mock instance
func NewMockPartyStore(ctrl *gomock.Controller) *MockPartyStore {
	mock := &MockPartyStore{ctrl: ctrl}
	mock.recorder = &MockPartyStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPartyStore) EXPECT() *MockPartyStoreMockRecorder {
	return m.recorder
}

// GetAll mocks base method
func (m *MockPartyStore) GetAll() ([]*proto.Party, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll")
	ret0, _ := ret[0].([]*proto.Party)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll
func (mr *MockPartyStoreMockRecorder) GetAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockPartyStore)(nil).GetAll))
}

// GetByID mocks base method
func (m *MockPartyStore) GetByID(arg0 string) (*proto.Party, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", arg0)
	ret0, _ := ret[0].(*proto.Party)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID
func (mr *MockPartyStoreMockRecorder) GetByID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockPartyStore)(nil).GetByID), arg0)
}

// Post mocks base method
func (m *MockPartyStore) Post(arg0 *proto.Party) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Post", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Post indicates an expected call of Post
func (mr *MockPartyStoreMockRecorder) Post(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Post", reflect.TypeOf((*MockPartyStore)(nil).Post), arg0)
}
