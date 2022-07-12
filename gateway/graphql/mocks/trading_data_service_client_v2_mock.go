// Code generated by MockGen. DO NOT EDIT.
// Source: code.vegaprotocol.io/data-node/gateway/graphql (interfaces: TradingDataServiceClientV2)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	v2 "code.vegaprotocol.io/protos/data-node/api/v2"
	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
)

// MockTradingDataServiceClientV2 is a mock of TradingDataServiceClientV2 interface.
type MockTradingDataServiceClientV2 struct {
	ctrl     *gomock.Controller
	recorder *MockTradingDataServiceClientV2MockRecorder
}

// MockTradingDataServiceClientV2MockRecorder is the mock recorder for MockTradingDataServiceClientV2.
type MockTradingDataServiceClientV2MockRecorder struct {
	mock *MockTradingDataServiceClientV2
}

// NewMockTradingDataServiceClientV2 creates a new mock instance.
func NewMockTradingDataServiceClientV2(ctrl *gomock.Controller) *MockTradingDataServiceClientV2 {
	mock := &MockTradingDataServiceClientV2{ctrl: ctrl}
	mock.recorder = &MockTradingDataServiceClientV2MockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTradingDataServiceClientV2) EXPECT() *MockTradingDataServiceClientV2MockRecorder {
	return m.recorder
}

// GetAssets mocks base method.
func (m *MockTradingDataServiceClientV2) GetAssets(arg0 context.Context, arg1 *v2.GetAssetsRequest, arg2 ...grpc.CallOption) (*v2.GetAssetsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetAssets", varargs...)
	ret0, _ := ret[0].(*v2.GetAssetsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAssets indicates an expected call of GetAssets.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetAssets(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAssets", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetAssets), varargs...)
}

// GetBalanceHistory mocks base method.
func (m *MockTradingDataServiceClientV2) GetBalanceHistory(arg0 context.Context, arg1 *v2.GetBalanceHistoryRequest, arg2 ...grpc.CallOption) (*v2.GetBalanceHistoryResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetBalanceHistory", varargs...)
	ret0, _ := ret[0].(*v2.GetBalanceHistoryResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBalanceHistory indicates an expected call of GetBalanceHistory.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetBalanceHistory(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBalanceHistory", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetBalanceHistory), varargs...)
}

// GetCandleData mocks base method.
func (m *MockTradingDataServiceClientV2) GetCandleData(arg0 context.Context, arg1 *v2.GetCandleDataRequest, arg2 ...grpc.CallOption) (*v2.GetCandleDataResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetCandleData", varargs...)
	ret0, _ := ret[0].(*v2.GetCandleDataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCandleData indicates an expected call of GetCandleData.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetCandleData(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCandleData", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetCandleData), varargs...)
}

// GetCandlesForMarket mocks base method.
func (m *MockTradingDataServiceClientV2) GetCandlesForMarket(arg0 context.Context, arg1 *v2.GetCandlesForMarketRequest, arg2 ...grpc.CallOption) (*v2.GetCandlesForMarketResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetCandlesForMarket", varargs...)
	ret0, _ := ret[0].(*v2.GetCandlesForMarketResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCandlesForMarket indicates an expected call of GetCandlesForMarket.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetCandlesForMarket(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCandlesForMarket", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetCandlesForMarket), varargs...)
}

// GetDeposits mocks base method.
func (m *MockTradingDataServiceClientV2) GetDeposits(arg0 context.Context, arg1 *v2.GetDepositsRequest, arg2 ...grpc.CallOption) (*v2.GetDepositsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetDeposits", varargs...)
	ret0, _ := ret[0].(*v2.GetDepositsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeposits indicates an expected call of GetDeposits.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetDeposits(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeposits", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetDeposits), varargs...)
}

// GetERC20ListAssetBundle mocks base method.
func (m *MockTradingDataServiceClientV2) GetERC20ListAssetBundle(arg0 context.Context, arg1 *v2.GetERC20ListAssetBundleRequest, arg2 ...grpc.CallOption) (*v2.GetERC20ListAssetBundleResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetERC20ListAssetBundle", varargs...)
	ret0, _ := ret[0].(*v2.GetERC20ListAssetBundleResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetERC20ListAssetBundle indicates an expected call of GetERC20ListAssetBundle.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetERC20ListAssetBundle(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetERC20ListAssetBundle", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetERC20ListAssetBundle), varargs...)
}

// GetERC20MultiSigSignerAddedBundles mocks base method.
func (m *MockTradingDataServiceClientV2) GetERC20MultiSigSignerAddedBundles(arg0 context.Context, arg1 *v2.GetERC20MultiSigSignerAddedBundlesRequest, arg2 ...grpc.CallOption) (*v2.GetERC20MultiSigSignerAddedBundlesResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetERC20MultiSigSignerAddedBundles", varargs...)
	ret0, _ := ret[0].(*v2.GetERC20MultiSigSignerAddedBundlesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetERC20MultiSigSignerAddedBundles indicates an expected call of GetERC20MultiSigSignerAddedBundles.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetERC20MultiSigSignerAddedBundles(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetERC20MultiSigSignerAddedBundles", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetERC20MultiSigSignerAddedBundles), varargs...)
}

// GetERC20MultiSigSignerRemovedBundles mocks base method.
func (m *MockTradingDataServiceClientV2) GetERC20MultiSigSignerRemovedBundles(arg0 context.Context, arg1 *v2.GetERC20MultiSigSignerRemovedBundlesRequest, arg2 ...grpc.CallOption) (*v2.GetERC20MultiSigSignerRemovedBundlesResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetERC20MultiSigSignerRemovedBundles", varargs...)
	ret0, _ := ret[0].(*v2.GetERC20MultiSigSignerRemovedBundlesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetERC20MultiSigSignerRemovedBundles indicates an expected call of GetERC20MultiSigSignerRemovedBundles.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetERC20MultiSigSignerRemovedBundles(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetERC20MultiSigSignerRemovedBundles", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetERC20MultiSigSignerRemovedBundles), varargs...)
}

// GetLiquidityProvisions mocks base method.
func (m *MockTradingDataServiceClientV2) GetLiquidityProvisions(arg0 context.Context, arg1 *v2.GetLiquidityProvisionsRequest, arg2 ...grpc.CallOption) (*v2.GetLiquidityProvisionsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetLiquidityProvisions", varargs...)
	ret0, _ := ret[0].(*v2.GetLiquidityProvisionsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLiquidityProvisions indicates an expected call of GetLiquidityProvisions.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetLiquidityProvisions(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLiquidityProvisions", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetLiquidityProvisions), varargs...)
}

// GetMarginLevels mocks base method.
func (m *MockTradingDataServiceClientV2) GetMarginLevels(arg0 context.Context, arg1 *v2.GetMarginLevelsRequest, arg2 ...grpc.CallOption) (*v2.GetMarginLevelsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetMarginLevels", varargs...)
	ret0, _ := ret[0].(*v2.GetMarginLevelsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMarginLevels indicates an expected call of GetMarginLevels.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetMarginLevels(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMarginLevels", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetMarginLevels), varargs...)
}

// GetMarketDataHistoryByID mocks base method.
func (m *MockTradingDataServiceClientV2) GetMarketDataHistoryByID(arg0 context.Context, arg1 *v2.GetMarketDataHistoryByIDRequest, arg2 ...grpc.CallOption) (*v2.GetMarketDataHistoryByIDResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetMarketDataHistoryByID", varargs...)
	ret0, _ := ret[0].(*v2.GetMarketDataHistoryByIDResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMarketDataHistoryByID indicates an expected call of GetMarketDataHistoryByID.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetMarketDataHistoryByID(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMarketDataHistoryByID", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetMarketDataHistoryByID), varargs...)
}

// GetMarkets mocks base method.
func (m *MockTradingDataServiceClientV2) GetMarkets(arg0 context.Context, arg1 *v2.GetMarketsRequest, arg2 ...grpc.CallOption) (*v2.GetMarketsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetMarkets", varargs...)
	ret0, _ := ret[0].(*v2.GetMarketsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMarkets indicates an expected call of GetMarkets.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetMarkets(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMarkets", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetMarkets), varargs...)
}

// GetNetworkLimits mocks base method.
func (m *MockTradingDataServiceClientV2) GetNetworkLimits(arg0 context.Context, arg1 *v2.GetNetworkLimitsRequest, arg2 ...grpc.CallOption) (*v2.GetNetworkLimitsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetNetworkLimits", varargs...)
	ret0, _ := ret[0].(*v2.GetNetworkLimitsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNetworkLimits indicates an expected call of GetNetworkLimits.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetNetworkLimits(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNetworkLimits", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetNetworkLimits), varargs...)
}

// GetOracleDataBySpecID mocks base method.
func (m *MockTradingDataServiceClientV2) GetOracleDataBySpecID(arg0 context.Context, arg1 *v2.GetOracleDataBySpecIDRequest, arg2 ...grpc.CallOption) (*v2.GetOracleDataBySpecIDResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetOracleDataBySpecID", varargs...)
	ret0, _ := ret[0].(*v2.GetOracleDataBySpecIDResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOracleDataBySpecID indicates an expected call of GetOracleDataBySpecID.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetOracleDataBySpecID(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOracleDataBySpecID", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetOracleDataBySpecID), varargs...)
}

// GetOracleDataConnection mocks base method.
func (m *MockTradingDataServiceClientV2) GetOracleDataConnection(arg0 context.Context, arg1 *v2.GetOracleDataConnectionRequest, arg2 ...grpc.CallOption) (*v2.GetOracleDataConnectionResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetOracleDataConnection", varargs...)
	ret0, _ := ret[0].(*v2.GetOracleDataConnectionResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOracleDataConnection indicates an expected call of GetOracleDataConnection.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetOracleDataConnection(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOracleDataConnection", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetOracleDataConnection), varargs...)
}

// GetOracleSpecByID mocks base method.
func (m *MockTradingDataServiceClientV2) GetOracleSpecByID(arg0 context.Context, arg1 *v2.GetOracleSpecByIDRequest, arg2 ...grpc.CallOption) (*v2.GetOracleSpecByIDResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetOracleSpecByID", varargs...)
	ret0, _ := ret[0].(*v2.GetOracleSpecByIDResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOracleSpecByID indicates an expected call of GetOracleSpecByID.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetOracleSpecByID(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOracleSpecByID", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetOracleSpecByID), varargs...)
}

// GetOracleSpecsConnection mocks base method.
func (m *MockTradingDataServiceClientV2) GetOracleSpecsConnection(arg0 context.Context, arg1 *v2.GetOracleSpecsConnectionRequest, arg2 ...grpc.CallOption) (*v2.GetOracleSpecsConnectionResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetOracleSpecsConnection", varargs...)
	ret0, _ := ret[0].(*v2.GetOracleSpecsConnectionResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOracleSpecsConnection indicates an expected call of GetOracleSpecsConnection.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetOracleSpecsConnection(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOracleSpecsConnection", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetOracleSpecsConnection), varargs...)
}

// GetOrder mocks base method.
func (m *MockTradingDataServiceClientV2) GetOrder(arg0 context.Context, arg1 *v2.GetOrderRequest, arg2 ...grpc.CallOption) (*v2.GetOrderResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetOrder", varargs...)
	ret0, _ := ret[0].(*v2.GetOrderResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrder indicates an expected call of GetOrder.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetOrder(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrder", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetOrder), varargs...)
}

// GetParties mocks base method.
func (m *MockTradingDataServiceClientV2) GetParties(arg0 context.Context, arg1 *v2.GetPartiesRequest, arg2 ...grpc.CallOption) (*v2.GetPartiesResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetParties", varargs...)
	ret0, _ := ret[0].(*v2.GetPartiesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetParties indicates an expected call of GetParties.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetParties(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetParties", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetParties), varargs...)
}

// GetPositionsByPartyConnection mocks base method.
func (m *MockTradingDataServiceClientV2) GetPositionsByPartyConnection(arg0 context.Context, arg1 *v2.GetPositionsByPartyConnectionRequest, arg2 ...grpc.CallOption) (*v2.GetPositionsByPartyConnectionResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetPositionsByPartyConnection", varargs...)
	ret0, _ := ret[0].(*v2.GetPositionsByPartyConnectionResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPositionsByPartyConnection indicates an expected call of GetPositionsByPartyConnection.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetPositionsByPartyConnection(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPositionsByPartyConnection", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetPositionsByPartyConnection), varargs...)
}

// GetRewardSummaries mocks base method.
func (m *MockTradingDataServiceClientV2) GetRewardSummaries(arg0 context.Context, arg1 *v2.GetRewardSummariesRequest, arg2 ...grpc.CallOption) (*v2.GetRewardSummariesResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetRewardSummaries", varargs...)
	ret0, _ := ret[0].(*v2.GetRewardSummariesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRewardSummaries indicates an expected call of GetRewardSummaries.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetRewardSummaries(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRewardSummaries", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetRewardSummaries), varargs...)
}

// GetRewards mocks base method.
func (m *MockTradingDataServiceClientV2) GetRewards(arg0 context.Context, arg1 *v2.GetRewardsRequest, arg2 ...grpc.CallOption) (*v2.GetRewardsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetRewards", varargs...)
	ret0, _ := ret[0].(*v2.GetRewardsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRewards indicates an expected call of GetRewards.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetRewards(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRewards", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetRewards), varargs...)
}

// GetTradesByMarket mocks base method.
func (m *MockTradingDataServiceClientV2) GetTradesByMarket(arg0 context.Context, arg1 *v2.GetTradesByMarketRequest, arg2 ...grpc.CallOption) (*v2.GetTradesByMarketResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetTradesByMarket", varargs...)
	ret0, _ := ret[0].(*v2.GetTradesByMarketResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTradesByMarket indicates an expected call of GetTradesByMarket.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetTradesByMarket(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTradesByMarket", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetTradesByMarket), varargs...)
}

// GetTradesByOrderID mocks base method.
func (m *MockTradingDataServiceClientV2) GetTradesByOrderID(arg0 context.Context, arg1 *v2.GetTradesByOrderIDRequest, arg2 ...grpc.CallOption) (*v2.GetTradesByOrderIDResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetTradesByOrderID", varargs...)
	ret0, _ := ret[0].(*v2.GetTradesByOrderIDResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTradesByOrderID indicates an expected call of GetTradesByOrderID.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetTradesByOrderID(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTradesByOrderID", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetTradesByOrderID), varargs...)
}

// GetTradesByParty mocks base method.
func (m *MockTradingDataServiceClientV2) GetTradesByParty(arg0 context.Context, arg1 *v2.GetTradesByPartyRequest, arg2 ...grpc.CallOption) (*v2.GetTradesByPartyResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetTradesByParty", varargs...)
	ret0, _ := ret[0].(*v2.GetTradesByPartyResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTradesByParty indicates an expected call of GetTradesByParty.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetTradesByParty(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTradesByParty", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetTradesByParty), varargs...)
}

// GetWithdrawals mocks base method.
func (m *MockTradingDataServiceClientV2) GetWithdrawals(arg0 context.Context, arg1 *v2.GetWithdrawalsRequest, arg2 ...grpc.CallOption) (*v2.GetWithdrawalsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetWithdrawals", varargs...)
	ret0, _ := ret[0].(*v2.GetWithdrawalsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWithdrawals indicates an expected call of GetWithdrawals.
func (mr *MockTradingDataServiceClientV2MockRecorder) GetWithdrawals(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWithdrawals", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).GetWithdrawals), varargs...)
}

// ListOracleData mocks base method.
func (m *MockTradingDataServiceClientV2) ListOracleData(arg0 context.Context, arg1 *v2.ListOracleDataRequest, arg2 ...grpc.CallOption) (*v2.ListOracleDataResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListOracleData", varargs...)
	ret0, _ := ret[0].(*v2.ListOracleDataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListOracleData indicates an expected call of ListOracleData.
func (mr *MockTradingDataServiceClientV2MockRecorder) ListOracleData(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListOracleData", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).ListOracleData), varargs...)
}

// ListOracleSpecs mocks base method.
func (m *MockTradingDataServiceClientV2) ListOracleSpecs(arg0 context.Context, arg1 *v2.ListOracleSpecsRequest, arg2 ...grpc.CallOption) (*v2.ListOracleSpecsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListOracleSpecs", varargs...)
	ret0, _ := ret[0].(*v2.ListOracleSpecsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListOracleSpecs indicates an expected call of ListOracleSpecs.
func (mr *MockTradingDataServiceClientV2MockRecorder) ListOracleSpecs(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListOracleSpecs", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).ListOracleSpecs), varargs...)
}

// ListOrderVersions mocks base method.
func (m *MockTradingDataServiceClientV2) ListOrderVersions(arg0 context.Context, arg1 *v2.ListOrderVersionsRequest, arg2 ...grpc.CallOption) (*v2.ListOrderVersionsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListOrderVersions", varargs...)
	ret0, _ := ret[0].(*v2.ListOrderVersionsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListOrderVersions indicates an expected call of ListOrderVersions.
func (mr *MockTradingDataServiceClientV2MockRecorder) ListOrderVersions(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListOrderVersions", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).ListOrderVersions), varargs...)
}

// ListOrders mocks base method.
func (m *MockTradingDataServiceClientV2) ListOrders(arg0 context.Context, arg1 *v2.ListOrdersRequest, arg2 ...grpc.CallOption) (*v2.ListOrdersResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListOrders", varargs...)
	ret0, _ := ret[0].(*v2.ListOrdersResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListOrders indicates an expected call of ListOrders.
func (mr *MockTradingDataServiceClientV2MockRecorder) ListOrders(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListOrders", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).ListOrders), varargs...)
}

// ListTransfers mocks base method.
func (m *MockTradingDataServiceClientV2) ListTransfers(arg0 context.Context, arg1 *v2.ListTransfersRequest, arg2 ...grpc.CallOption) (*v2.ListTransfersResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListTransfers", varargs...)
	ret0, _ := ret[0].(*v2.ListTransfersResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListTransfers indicates an expected call of ListTransfers.
func (mr *MockTradingDataServiceClientV2MockRecorder) ListTransfers(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListTransfers", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).ListTransfers), varargs...)
}

// ListVotes mocks base method.
func (m *MockTradingDataServiceClientV2) ListVotes(arg0 context.Context, arg1 *v2.ListVotesRequest, arg2 ...grpc.CallOption) (*v2.ListVotesResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListVotes", varargs...)
	ret0, _ := ret[0].(*v2.ListVotesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListVotes indicates an expected call of ListVotes.
func (mr *MockTradingDataServiceClientV2MockRecorder) ListVotes(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListVotes", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).ListVotes), varargs...)
}

// MarketsDataSubscribe mocks base method.
func (m *MockTradingDataServiceClientV2) MarketsDataSubscribe(arg0 context.Context, arg1 *v2.MarketsDataSubscribeRequest, arg2 ...grpc.CallOption) (v2.TradingDataService_MarketsDataSubscribeClient, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "MarketsDataSubscribe", varargs...)
	ret0, _ := ret[0].(v2.TradingDataService_MarketsDataSubscribeClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MarketsDataSubscribe indicates an expected call of MarketsDataSubscribe.
func (mr *MockTradingDataServiceClientV2MockRecorder) MarketsDataSubscribe(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarketsDataSubscribe", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).MarketsDataSubscribe), varargs...)
}

// ObserveVotes mocks base method.
func (m *MockTradingDataServiceClientV2) ObserveVotes(arg0 context.Context, arg1 *v2.ObserveVotesRequest, arg2 ...grpc.CallOption) (v2.TradingDataService_ObserveVotesClient, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ObserveVotes", varargs...)
	ret0, _ := ret[0].(v2.TradingDataService_ObserveVotesClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObserveVotes indicates an expected call of ObserveVotes.
func (mr *MockTradingDataServiceClientV2MockRecorder) ObserveVotes(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObserveVotes", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).ObserveVotes), varargs...)
}

// SubscribeToCandleData mocks base method.
func (m *MockTradingDataServiceClientV2) SubscribeToCandleData(arg0 context.Context, arg1 *v2.SubscribeToCandleDataRequest, arg2 ...grpc.CallOption) (v2.TradingDataService_SubscribeToCandleDataClient, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SubscribeToCandleData", varargs...)
	ret0, _ := ret[0].(v2.TradingDataService_SubscribeToCandleDataClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SubscribeToCandleData indicates an expected call of SubscribeToCandleData.
func (mr *MockTradingDataServiceClientV2MockRecorder) SubscribeToCandleData(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubscribeToCandleData", reflect.TypeOf((*MockTradingDataServiceClientV2)(nil).SubscribeToCandleData), varargs...)
}
