package gql

import (
	"context"
	"strconv"

	vega "code.vegaprotocol.io/protos/vega/api/v1"
)

type statisticsResolver VegaResolverRoot

func (s *statisticsResolver) BlockHeight(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.BlockHeight, 10), nil
}

func (s *statisticsResolver) BacklogLength(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.BacklogLength, 10), nil
}

func (s *statisticsResolver) TotalPeers(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.TotalPeers, 10), nil
}

func (s *statisticsResolver) TxPerBlock(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.TxPerBlock, 10), nil
}

func (s *statisticsResolver) AverageTxBytes(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.AverageTxBytes, 10), nil
}

func (s *statisticsResolver) AverageOrdersPerBlock(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.AverageOrdersPerBlock, 10), nil
}

func (s *statisticsResolver) TradesPerSecond(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.TradesPerSecond, 10), nil
}

func (s *statisticsResolver) OrdersPerSecond(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.OrdersPerSecond, 10), nil
}

func (s *statisticsResolver) TotalMarkets(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.TotalMarkets, 10), nil
}

func (s *statisticsResolver) TotalAmendOrder(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.TotalAmendOrder, 10), nil
}

func (s *statisticsResolver) TotalCancelOrder(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.TotalCancelOrder, 10), nil
}

func (s *statisticsResolver) TotalCreateOrder(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.TotalCreateOrder, 10), nil
}

func (s *statisticsResolver) TotalOrders(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.TotalOrders, 10), nil
}

func (s *statisticsResolver) TotalTrades(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.TotalTrades, 10), nil
}

func (s *statisticsResolver) OrderSubscriptions(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(uint64(obj.OrderSubscriptions), 10), nil
}

func (s *statisticsResolver) TradeSubscriptions(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(uint64(obj.TradeSubscriptions), 10), nil
}

func (s *statisticsResolver) CandleSubscriptions(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(uint64(obj.CandleSubscriptions), 10), nil
}

func (s *statisticsResolver) MarketDepthSubscriptions(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(uint64(obj.MarketDepthSubscriptions), 10), nil
}

func (s *statisticsResolver) PositionsSubscriptions(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(uint64(obj.PositionsSubscriptions), 10), nil
}

func (s *statisticsResolver) AccountSubscriptions(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(uint64(obj.AccountSubscriptions), 10), nil
}

func (s *statisticsResolver) MarketDataSubscriptions(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(uint64(obj.MarketDataSubscriptions), 10), nil
}

func (s *statisticsResolver) MarketDepthUpdatesSubscriptions(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(uint64(obj.MarketDepthUpdatesSubscriptions), 10), nil
}

func (s *statisticsResolver) BlockDuration(ctx context.Context, obj *vega.Statistics) (string, error) {
	return strconv.FormatUint(obj.BlockDuration, 10), nil
}

func (s *statisticsResolver) Status(ctx context.Context, obj *vega.Statistics) (string, error) {
	return obj.Status.String(), nil
}
