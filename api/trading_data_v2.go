package api

import (
	"context"
	"fmt"
	"time"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/sqlstore"
	v2 "code.vegaprotocol.io/protos/data-node/api/v2"
	types "code.vegaprotocol.io/protos/vega"
)

type tradingDataServiceV2 struct {
	v2.UnimplementedTradingDataServiceServer
	balanceStore    *sqlstore.Balances
	marketDataStore *sqlstore.MarketData
}

func (bs *tradingDataServiceV2) QueryBalanceHistory(ctx context.Context, req *v2.QueryBalanceHistoryRequest) (*v2.QueryBalanceHistoryResponse, error) {
	if bs.balanceStore == nil {
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

	balances, err := bs.balanceStore.Query(filter, groupBy)
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

func entityMarketDataListToProtoList(list []entities.MarketData) []*types.MarketData {
	if len(list) == 0 {
		return nil
	}

	results := make([]*types.MarketData, 0, len(list))

	for _, item := range list {
		results = append(results, item.ToProto())
	}

	return results
}

func (bs *tradingDataServiceV2) GetMarketDataHistoryByID(ctx context.Context, req *v2.GetMarketDataHistoryByIDRequest) (*v2.GetMarketDataHistoryByIDResponse, error) {
	var startTime, endTime time.Time

	if req.StartTimestamp != nil {
		startTime = time.Unix(0, *req.StartTimestamp)
	}

	if req.EndTimestamp != nil {
		endTime = time.Unix(0, *req.EndTimestamp)
	}

	pagination := sqlstore.Pagination{
		Offset:     req.Pagination.Skip,
		Limit:      req.Pagination.Limit,
		Descending: req.Pagination.Descending,
	}

	if req.StartTimestamp != nil && req.EndTimestamp != nil {
		return bs.getMarketDataHistoryByID(ctx, req.MarketId, startTime, endTime, pagination)
	}

	if req.StartTimestamp != nil {
		return bs.getMarketDataHistoryFromDateByID(ctx, req.MarketId, startTime, pagination)
	}

	if req.EndTimestamp != nil {
		return bs.getMarketDataHistoryToDateByID(ctx, req.MarketId, endTime, pagination)
	}

	return bs.getMarketDataByID(ctx, req.MarketId)
}

func parseMarketDataResults(results []entities.MarketData) (*v2.GetMarketDataHistoryByIDResponse, error) {
	response := v2.GetMarketDataHistoryByIDResponse{
		MarketData: entityMarketDataListToProtoList(results),
	}

	return &response, nil
}

func (bs *tradingDataServiceV2) getMarketDataHistoryByID(ctx context.Context, id string, start, end time.Time, pagination sqlstore.Pagination) (*v2.GetMarketDataHistoryByIDResponse, error) {
	results, err := bs.marketDataStore.GetBetweenDatesByID(ctx, id, start, end, pagination)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve market data history for market id: %w", err)
	}

	return parseMarketDataResults(results)
}

func (bs *tradingDataServiceV2) getMarketDataByID(ctx context.Context, id string) (*v2.GetMarketDataHistoryByIDResponse, error) {
	results, err := bs.marketDataStore.GetByID(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve market data history for market id: %w", err)
	}

	return parseMarketDataResults([]entities.MarketData{results})
}

func (bs *tradingDataServiceV2) getMarketDataHistoryFromDateByID(ctx context.Context, id string, start time.Time, pagination sqlstore.Pagination) (*v2.GetMarketDataHistoryByIDResponse, error) {
	results, err := bs.marketDataStore.GetFromDateByID(ctx, id, start, pagination)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve market data history for market id: %w", err)
	}

	return parseMarketDataResults(results)
}

func (bs *tradingDataServiceV2) getMarketDataHistoryToDateByID(ctx context.Context, id string, end time.Time, pagination sqlstore.Pagination) (*v2.GetMarketDataHistoryByIDResponse, error) {
	results, err := bs.marketDataStore.GetToDateByID(ctx, id, end, pagination)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve market data history for market id: %w", err)
	}

	return parseMarketDataResults(results)
}
