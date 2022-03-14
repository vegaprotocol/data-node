package api

import (
	"code.vegaprotocol.io/data-node/vegatime"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/sqlstore"
	v2 "code.vegaprotocol.io/protos/data-node/api/v2"
	"code.vegaprotocol.io/protos/vega"
	"google.golang.org/grpc/codes"
)

var defaultPaginationV2 = entities.Pagination{
	Skip:       0,
	Limit:      1000,
	Descending: true,
}

type tradingDataServiceV2 struct {
	v2.UnimplementedTradingDataServiceServer
	balanceStore       *sqlstore.Balances
	orderStore         *sqlstore.Orders
	networkLimitsStore *sqlstore.NetworkLimits
	marketDataStore    *sqlstore.MarketData
	tradeStore         *sqlstore.Trades
	candleStore        *sqlstore.Candles
}

func (t *tradingDataServiceV2) QueryBalanceHistory(ctx context.Context, req *v2.QueryBalanceHistoryRequest) (*v2.QueryBalanceHistoryResponse, error) {
	if t.balanceStore == nil {
		return nil, fmt.Errorf("sql balance store not available")
	}

	filter, err := entities.AccountFilterFromProto(req.Filter)
	if err != nil {
		return nil, fmt.Errorf("parsing filter: %w", err)
	}

	groupBy := []entities.AccountField{}
	for _, field := range req.GroupBy {
		field, err := entities.AccountFieldFromProto(field)
		if err != nil {
			return nil, fmt.Errorf("parsing group by list: %w", err)
		}
		groupBy = append(groupBy, field)
	}

	balances, err := t.balanceStore.Query(filter, groupBy)
	if err != nil {
		return nil, fmt.Errorf("querying balances: %w", err)
	}

	pbBalances := make([]*v2.AggregatedBalance, len(*balances))
	for i, balance := range *balances {
		pbBalance := balance.ToProto()
		pbBalances[i] = &pbBalance
	}

	return &v2.QueryBalanceHistoryResponse{Balances: pbBalances}, nil
}

func (t *tradingDataServiceV2) OrdersByMarket(ctx context.Context, req *v2.OrdersByMarketRequest) (*v2.OrdersByMarketResponse, error) {
	if t.orderStore == nil {
		return nil, errors.New("sql order store not available")
	}

	p := defaultPaginationV2
	if req.Pagination != nil {
		p = entities.PaginationFromProto(req.Pagination)
	}

	orders, err := t.orderStore.GetByMarket(ctx, req.MarketId, p)
	if err != nil {
		return nil, apiError(codes.InvalidArgument, ErrOrderServiceGetByParty, err)
	}

	pbOrders := make([]*vega.Order, len(orders))
	for i, order := range orders {
		pbOrders[i] = order.ToProto()
	}

	return &v2.OrdersByMarketResponse{
		Orders: pbOrders,
	}, nil
}

func entityMarketDataListToProtoList(list []entities.MarketData) []*vega.MarketData {
	if len(list) == 0 {
		return nil
	}

	results := make([]*vega.MarketData, 0, len(list))

	for _, item := range list {
		results = append(results, item.ToProto())
	}

	return results
}

func (t *tradingDataServiceV2) GetMarketDataHistoryByID(ctx context.Context, req *v2.GetMarketDataHistoryByIDRequest) (*v2.GetMarketDataHistoryByIDResponse, error) {
	if t.marketDataStore == nil {
		return nil, errors.New("sql market data store not available")
	}

	var startTime, endTime time.Time

	if req.StartTimestamp != nil {
		startTime = time.Unix(0, *req.StartTimestamp)
	}

	if req.EndTimestamp != nil {
		endTime = time.Unix(0, *req.EndTimestamp)
	}

	pagination := defaultPaginationV2
	if req.Pagination != nil {
		pagination = entities.PaginationFromProto(req.Pagination)
	}

	if req.StartTimestamp != nil && req.EndTimestamp != nil {
		return t.getMarketDataHistoryByID(ctx, req.MarketId, startTime, endTime, pagination)
	}

	if req.StartTimestamp != nil {
		return t.getMarketDataHistoryFromDateByID(ctx, req.MarketId, startTime, pagination)
	}

	if req.EndTimestamp != nil {
		return t.getMarketDataHistoryToDateByID(ctx, req.MarketId, endTime, pagination)
	}

	return t.getMarketDataByID(ctx, req.MarketId)
}

func parseMarketDataResults(results []entities.MarketData) (*v2.GetMarketDataHistoryByIDResponse, error) {
	response := v2.GetMarketDataHistoryByIDResponse{
		MarketData: entityMarketDataListToProtoList(results),
	}

	return &response, nil
}

func (t *tradingDataServiceV2) getMarketDataHistoryByID(ctx context.Context, id string, start, end time.Time, pagination entities.Pagination) (*v2.GetMarketDataHistoryByIDResponse, error) {
	results, err := t.marketDataStore.GetBetweenDatesByID(ctx, id, start, end, pagination)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve market data history for market id: %w", err)
	}

	return parseMarketDataResults(results)
}

func (t *tradingDataServiceV2) getMarketDataByID(ctx context.Context, id string) (*v2.GetMarketDataHistoryByIDResponse, error) {
	results, err := t.marketDataStore.GetMarketDataByID(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve market data history for market id: %w", err)
	}

	return parseMarketDataResults([]entities.MarketData{results})
}

func (t *tradingDataServiceV2) getMarketDataHistoryFromDateByID(ctx context.Context, id string, start time.Time, pagination entities.Pagination) (*v2.GetMarketDataHistoryByIDResponse, error) {
	results, err := t.marketDataStore.GetFromDateByID(ctx, id, start, pagination)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve market data history for market id: %w", err)
	}

	return parseMarketDataResults(results)
}

func (t *tradingDataServiceV2) getMarketDataHistoryToDateByID(ctx context.Context, id string, end time.Time, pagination entities.Pagination) (*v2.GetMarketDataHistoryByIDResponse, error) {
	results, err := t.marketDataStore.GetToDateByID(ctx, id, end, pagination)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve market data history for market id: %w", err)
	}

	return parseMarketDataResults(results)
}

func (t *tradingDataServiceV2) GetNetworkLimits(ctx context.Context, req *v2.GetNetworkLimitsRequest) (*v2.GetNetworkLimitsResponse, error) {
	if t.networkLimitsStore == nil {
		return nil, errors.New("sql network limits store is not available")
	}

	limits, err := t.networkLimitsStore.GetLatest(ctx)
	if err != nil {
		return nil, apiError(codes.Unknown, ErrGetNetworkLimits, err)
	}

	return &v2.GetNetworkLimitsResponse{Limits: limits.ToProto()}, nil
}

// Candles for a given market, time range and interval.  Interval must be a valid postgres interval value
func (t *tradingDataServiceV2) Candles(ctx context.Context, request *v2.CandlesRequest) (*v2.CandlesResponse, error) {

	marketId, err := hex.DecodeString(request.MarketId)
	if err != nil {
		return nil, apiError(codes.InvalidArgument, ErrCandleServiceGetCandles, fmt.Errorf("market id is invalid:%w", err))
	}

	err = t.candleStore.ValidateInterval(ctx, request.Interval)
	if err != nil {
		return nil, apiError(codes.InvalidArgument, ErrCandleServiceGetCandles, fmt.Errorf("interval is invalid:%w", err))
	}

	from := vegatime.UnixNano(request.FromTimestamp)
	to := vegatime.UnixNano(request.ToTimestamp)

	pagination := defaultPaginationV2
	if request.Pagination != nil {
		pagination = entities.PaginationFromProto(request.Pagination)
	}

	candles, err := t.candleStore.GetCandlesForInterval(ctx, request.Interval, &from, &to, marketId, pagination)
	if err != nil {
		return nil, fmt.Errorf("getting candles for interval:%w", err)
	}

	var protoCandles []*v2.Candle
	for _, candle := range candles {
		protoCandles = append(protoCandles, candle.ToV2CandleProto())
	}

	return &v2.CandlesResponse{Candles: protoCandles}, nil
}

// CandlesSubscribe subscribes to candle updates for a given market and interval.  Interval must be a valid postgres interval value
func (t *tradingDataServiceV2) CandlesSubscribe(req *v2.CandlesSubscribeRequest, srv v2.TradingDataService_CandlesSubscribeServer) error {

	ctx, cancel := context.WithCancel(srv.Context())
	defer cancel()

	marketId, err := hex.DecodeString(req.MarketId)
	if err != nil {
		return apiError(codes.InvalidArgument, ErrCandleServiceGetCandles, fmt.Errorf("market id is invalid:%w", err))
	}

	err = t.candleStore.ValidateInterval(ctx, req.Interval)
	if err != nil {
		return apiError(codes.InvalidArgument, ErrCandleServiceSubscribeToCandles, fmt.Errorf("interval is invalid:%w", err))
	}

	subscriptionId, candlesChan, err := t.tradeStore.SubscribeToCandle(ctx, marketId, req.Interval)
	if err != nil {
		return fmt.Errorf("subscribing to candles:%w", err)
	}

	for {
		select {
		case candle, ok := <-candlesChan:

			if !ok {
				return fmt.Errorf("channel closed")
			}

			resp := &v2.CandlesSubscribeResponse{
				Candle: candle.ToV2CandleProto(),
			}
			if err = srv.Send(resp); err != nil {
				return fmt.Errorf("sending candles:%w", err)
			}
		case <-ctx.Done():
			t.tradeStore.UnsubscribeFromCandle(subscriptionId)
			err = ctx.Err()
			if err != nil {
				return fmt.Errorf("context done:%w", err)
			}
			return nil
		}

	}

	return nil
}
