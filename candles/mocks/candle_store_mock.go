// Code generated by MockGen. DO NOT EDIT.
// Source: code.vegaprotocol.io/data-node/candles (interfaces: CandleStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	proto "code.vegaprotocol.io/data-node/proto/vega"
	storage "code.vegaprotocol.io/data-node/storage"
	context "context"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// MockCandleStore is a mock of CandleStore interface
type MockCandleStore struct {
	ctrl     *gomock.Controller
	recorder *MockCandleStoreMockRecorder
}

// MockCandleStoreMockRecorder is the mock recorder for MockCandleStore
type MockCandleStoreMockRecorder struct {
	mock *MockCandleStore
}

// NewMockCandleStore creates a new mock instance
func NewMockCandleStore(ctrl *gomock.Controller) *MockCandleStore {
	mock := &MockCandleStore{ctrl: ctrl}
	mock.recorder = &MockCandleStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCandleStore) EXPECT() *MockCandleStoreMockRecorder {
	return m.recorder
}

// GetCandles mocks base method
func (m *MockCandleStore) GetCandles(arg0 context.Context, arg1 string, arg2 time.Time, arg3 proto.Interval) ([]*proto.Candle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCandles", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*proto.Candle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCandles indicates an expected call of GetCandles
func (mr *MockCandleStoreMockRecorder) GetCandles(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCandles", reflect.TypeOf((*MockCandleStore)(nil).GetCandles), arg0, arg1, arg2, arg3)
}

// Subscribe mocks base method
func (m *MockCandleStore) Subscribe(arg0 *storage.InternalTransport) uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe", arg0)
	ret0, _ := ret[0].(uint64)
	return ret0
}

// Subscribe indicates an expected call of Subscribe
func (mr *MockCandleStoreMockRecorder) Subscribe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockCandleStore)(nil).Subscribe), arg0)
}

// Unsubscribe mocks base method
func (m *MockCandleStore) Unsubscribe(arg0 uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unsubscribe", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unsubscribe indicates an expected call of Unsubscribe
func (mr *MockCandleStoreMockRecorder) Unsubscribe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unsubscribe", reflect.TypeOf((*MockCandleStore)(nil).Unsubscribe), arg0)
}
