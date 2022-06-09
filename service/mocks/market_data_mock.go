// Code generated by MockGen. DO NOT EDIT.
// Source: code.vegaprotocol.io/data-node/service (interfaces: MarketDataStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	time "time"

	entities "code.vegaprotocol.io/data-node/entities"
	gomock "github.com/golang/mock/gomock"
)

// MockMarketDataStore is a mock of MarketDataStore interface.
type MockMarketDataStore struct {
	ctrl     *gomock.Controller
	recorder *MockMarketDataStoreMockRecorder
}

// MockMarketDataStoreMockRecorder is the mock recorder for MockMarketDataStore.
type MockMarketDataStoreMockRecorder struct {
	mock *MockMarketDataStore
}

// NewMockMarketDataStore creates a new mock instance.
func NewMockMarketDataStore(ctrl *gomock.Controller) *MockMarketDataStore {
	mock := &MockMarketDataStore{ctrl: ctrl}
	mock.recorder = &MockMarketDataStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMarketDataStore) EXPECT() *MockMarketDataStoreMockRecorder {
	return m.recorder
}

// Add mocks base method.
func (m *MockMarketDataStore) Add(arg0 *entities.MarketData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add.
func (mr *MockMarketDataStoreMockRecorder) Add(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockMarketDataStore)(nil).Add), arg0)
}

// Flush mocks base method.
func (m *MockMarketDataStore) Flush(arg0 context.Context) ([]*entities.MarketData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Flush", arg0)
	ret0, _ := ret[0].([]*entities.MarketData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Flush indicates an expected call of Flush.
func (mr *MockMarketDataStoreMockRecorder) Flush(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Flush", reflect.TypeOf((*MockMarketDataStore)(nil).Flush), arg0)
}

// GetBetweenDatesByID mocks base method.
func (m *MockMarketDataStore) GetBetweenDatesByID(arg0 context.Context, arg1 string, arg2, arg3 time.Time, arg4 entities.OffsetPagination) ([]entities.MarketData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBetweenDatesByID", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].([]entities.MarketData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBetweenDatesByID indicates an expected call of GetBetweenDatesByID.
func (mr *MockMarketDataStoreMockRecorder) GetBetweenDatesByID(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBetweenDatesByID", reflect.TypeOf((*MockMarketDataStore)(nil).GetBetweenDatesByID), arg0, arg1, arg2, arg3, arg4)
}

// GetFromDateByID mocks base method.
func (m *MockMarketDataStore) GetFromDateByID(arg0 context.Context, arg1 string, arg2 time.Time, arg3 entities.OffsetPagination) ([]entities.MarketData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFromDateByID", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]entities.MarketData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFromDateByID indicates an expected call of GetFromDateByID.
func (mr *MockMarketDataStoreMockRecorder) GetFromDateByID(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFromDateByID", reflect.TypeOf((*MockMarketDataStore)(nil).GetFromDateByID), arg0, arg1, arg2, arg3)
}

// GetMarketDataByID mocks base method.
func (m *MockMarketDataStore) GetMarketDataByID(arg0 context.Context, arg1 string) (entities.MarketData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMarketDataByID", arg0, arg1)
	ret0, _ := ret[0].(entities.MarketData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMarketDataByID indicates an expected call of GetMarketDataByID.
func (mr *MockMarketDataStoreMockRecorder) GetMarketDataByID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMarketDataByID", reflect.TypeOf((*MockMarketDataStore)(nil).GetMarketDataByID), arg0, arg1)
}

// GetMarketsData mocks base method.
func (m *MockMarketDataStore) GetMarketsData(arg0 context.Context) ([]entities.MarketData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMarketsData", arg0)
	ret0, _ := ret[0].([]entities.MarketData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMarketsData indicates an expected call of GetMarketsData.
func (mr *MockMarketDataStoreMockRecorder) GetMarketsData(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMarketsData", reflect.TypeOf((*MockMarketDataStore)(nil).GetMarketsData), arg0)
}

// GetToDateByID mocks base method.
func (m *MockMarketDataStore) GetToDateByID(arg0 context.Context, arg1 string, arg2 time.Time, arg3 entities.OffsetPagination) ([]entities.MarketData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetToDateByID", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]entities.MarketData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetToDateByID indicates an expected call of GetToDateByID.
func (mr *MockMarketDataStoreMockRecorder) GetToDateByID(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetToDateByID", reflect.TypeOf((*MockMarketDataStore)(nil).GetToDateByID), arg0, arg1, arg2, arg3)
}