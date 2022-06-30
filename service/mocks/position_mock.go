// Code generated by MockGen. DO NOT EDIT.
// Source: code.vegaprotocol.io/data-node/service (interfaces: PositionStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	entities "code.vegaprotocol.io/data-node/entities"
	context "context"
	v2 "code.vegaprotocol.io/protos/data-node/api/v2"
)

// MockPositionStore is a mock of PositionStore interface.
type MockPositionStore struct {
	ctrl     *gomock.Controller
	recorder *MockPositionStoreMockRecorder
}

// MockPositionStoreMockRecorder is the mock recorder for MockPositionStore.
type MockPositionStoreMockRecorder struct {
	mock *MockPositionStore
}

// NewMockPositionStore creates a new mock instance.
func NewMockPositionStore(ctrl *gomock.Controller) *MockPositionStore {
	mock := &MockPositionStore{ctrl: ctrl}
	mock.recorder = &MockPositionStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPositionStore) EXPECT() *MockPositionStoreMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockPositionStore) Add(arg0 context.Context, arg1 entities.Position) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockPositionStoreMockRecorder) Add(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockPositionStore)(nil).Add), arg0, arg1)
}

// Flush mocks base method.
func (m *MockPositionStore) Flush(arg0 context.Context) ([]entities.Position, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Flush", arg0)
	ret0, _ := ret[0].([]entities.Position)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Flush indicates an expected call of Flush.
func (mr *MockPositionStoreMockRecorder) Flush(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Flush", reflect.TypeOf((*MockPositionStore)(nil).Flush), arg0)
}

// GetAll mocks base method.
func (m *MockPositionStore) GetAll(arg0 context.Context) ([]entities.Position, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", arg0)
	ret0, _ := ret[0].([]entities.Position)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockPositionStoreMockRecorder) GetAll(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockPositionStore)(nil).GetAll), arg0)
}

// GetByMarket mocks base method.
func (m *MockPositionStore) GetByMarket(arg0 context.Context, arg1 entities.MarketID) ([]entities.Position, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByMarket", arg0, arg1)
	ret0, _ := ret[0].([]entities.Position)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByMarket indicates an expected call of GetByMarket.
func (mr *MockPositionStoreMockRecorder) GetByMarket(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByMarket", reflect.TypeOf((*MockPositionStore)(nil).GetByMarket), arg0, arg1)
}

// GetByMarketAndParty mocks base method.
func (m *MockPositionStore) GetByMarketAndParty(arg0 context.Context, arg1 entities.MarketID, arg2 entities.PartyID) (entities.Position, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByMarketAndParty", arg0, arg1, arg2)
	ret0, _ := ret[0].(entities.Position)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByMarketAndParty indicates an expected call of GetByMarketAndParty.
func (mr *MockPositionStoreMockRecorder) GetByMarketAndParty(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByMarketAndParty", reflect.TypeOf((*MockPositionStore)(nil).GetByMarketAndParty), arg0, arg1, arg2)
}

// GetByParty mocks base method.
func (m *MockPositionStore) GetByParty(arg0 context.Context, arg1 entities.PartyID) ([]entities.Position, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByParty", arg0, arg1)
	ret0, _ := ret[0].([]entities.Position)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByParty indicates an expected call of GetByParty.
func (mr *MockPositionStoreMockRecorder) GetByParty(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByParty", reflect.TypeOf((*MockPositionStore)(nil).GetByParty), arg0, arg1)
}

// GetByPartyConnection mocks base method.
func (m *MockPositionStore) GetByPartyConnection(arg0 context.Context, arg1 entities.PartyID, arg2 entities.MarketID, arg3 entities.CursorPagination) entities.ConnectionData[*v2.PositionEdge,entities.Position] {
m.ctrl.T.Helper()
ret := m.ctrl.Call(m, "GetByPartyConnection", arg0, arg1, arg2, arg3)
ret0, _ := ret[0].(entities.ConnectionData[*v2.PositionEdge,entities.Position])
return ret0
}

// GetByPartyConnection indicates an expected call of GetByPartyConnection.
func (mr *MockPositionStoreMockRecorder) GetByPartyConnection(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByPartyConnection", reflect.TypeOf((*MockPositionStore)(nil).GetByPartyConnection), arg0, arg1, arg2, arg3)
}
