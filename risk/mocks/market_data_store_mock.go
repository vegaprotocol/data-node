// Code generated by MockGen. DO NOT EDIT.
// Source: code.vegaprotocol.io/data-node/risk (interfaces: MarketDataStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	proto "code.vegaprotocol.io/data-node/proto/vega"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockMarketDataStore is a mock of MarketDataStore interface
type MockMarketDataStore struct {
	ctrl     *gomock.Controller
	recorder *MockMarketDataStoreMockRecorder
}

// MockMarketDataStoreMockRecorder is the mock recorder for MockMarketDataStore
type MockMarketDataStoreMockRecorder struct {
	mock *MockMarketDataStore
}

// NewMockMarketDataStore creates a new mock instance
func NewMockMarketDataStore(ctrl *gomock.Controller) *MockMarketDataStore {
	mock := &MockMarketDataStore{ctrl: ctrl}
	mock.recorder = &MockMarketDataStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMarketDataStore) EXPECT() *MockMarketDataStoreMockRecorder {
	return m.recorder
}

// GetByID mocks base method
func (m *MockMarketDataStore) GetByID(arg0 string) (proto.MarketData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", arg0)
	ret0, _ := ret[0].(proto.MarketData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID
func (mr *MockMarketDataStoreMockRecorder) GetByID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockMarketDataStore)(nil).GetByID), arg0)
}
