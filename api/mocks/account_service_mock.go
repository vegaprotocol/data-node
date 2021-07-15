// Code generated by MockGen. DO NOT EDIT.
// Source: code.vegaprotocol.io/data-node/api (interfaces: AccountService)

// Package mocks is a generated GoMock package.
package mocks

import (
	v1 "code.vegaprotocol.io/data-node/proto/commands/v1"
	context "context"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockAccountService is a mock of AccountService interface
type MockAccountService struct {
	ctrl     *gomock.Controller
	recorder *MockAccountServiceMockRecorder
}

// MockAccountServiceMockRecorder is the mock recorder for MockAccountService
type MockAccountServiceMockRecorder struct {
	mock *MockAccountService
}

// NewMockAccountService creates a new mock instance
func NewMockAccountService(ctrl *gomock.Controller) *MockAccountService {
	mock := &MockAccountService{ctrl: ctrl}
	mock.recorder = &MockAccountServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAccountService) EXPECT() *MockAccountServiceMockRecorder {
	return m.recorder
}

// PrepareWithdraw mocks base method
func (m *MockAccountService) PrepareWithdraw(arg0 context.Context, arg1 *v1.WithdrawSubmission) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PrepareWithdraw", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PrepareWithdraw indicates an expected call of PrepareWithdraw
func (mr *MockAccountServiceMockRecorder) PrepareWithdraw(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PrepareWithdraw", reflect.TypeOf((*MockAccountService)(nil).PrepareWithdraw), arg0, arg1)
}
