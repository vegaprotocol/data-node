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

func (bs *tradingDataServiceV2) GetMarketDataByID(ctx context.Context, req *v2.MarketDataByIDRequest) (*v2.MarketDataByIDResponse, error) {
	results, err := bs.marketDataStore.GetByID(ctx, req.MarketId)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve market data for given market ID: %w", err)
	}

	response := &v2.MarketDataByIDResponse{
		MarketData: results.ToProto(),
	}

	return response, nil
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

func (bs *tradingDataServiceV2) GetMarketsData(ctx context.Context, _ *v2.MarketsDataRequest) (*v2.MarketsDataResponse, error) {
	results, err := bs.marketDataStore.GetAll(ctx)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve latest market data snapshot: %w", err)
	}

	response := v2.MarketsDataResponse{
		MarketsData: entityMarketDataListToProtoList(results),
	}

	return &response, nil
}

func (bs *tradingDataServiceV2) GetMarketDataHistoryByID(ctx context.Context, req *v2.MarketDataHistoryByIDRequest) (*v2.MarketDataHistoryByIDResponse, error) {
	startTime := time.Unix(0, req.StartTimestamp)
	endTime := time.Unix(0, req.EndTimestamp)

	results, err := bs.marketDataStore.GetBetweenDatesByID(ctx, req.MarketId, startTime, endTime)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve market data history for market id: %w", err)
	}

	response := v2.MarketDataHistoryByIDResponse{
		MarketData: entityMarketDataListToProtoList(results),
	}

	return &response, nil
}

func (bs *tradingDataServiceV2) GetMarketDataHistoryFromDateByID(ctx context.Context, req *v2.MarketDataHistoryFromDateByIDRequest) (*v2.MarketDataHistoryFromDateByIDResponse, error) {
	from := time.Unix(0, req.StartTimestamp)
	results, err := bs.marketDataStore.GetFromDateByID(ctx, req.MarketId, from)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve market data history for market id: %w", err)
	}

	response := v2.MarketDataHistoryFromDateByIDResponse{
		MarketData: entityMarketDataListToProtoList(results),
	}

	return &response, nil
}

func (bs *tradingDataServiceV2) GetMarketDataHistoryToDateByID(ctx context.Context, req *v2.MarketDataHistoryToDateByIDRequest) (*v2.MarketDataHistoryToDateByIDResponse, error) {
	to := time.Unix(0, req.EndTimestamp)
	results, err := bs.marketDataStore.GetToDateByID(ctx, req.MarketId, to)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve market data history for market id: %w", err)
	}

	response := v2.MarketDataHistoryToDateByIDResponse{
		MarketData: entityMarketDataListToProtoList(results),
	}

	return &response, nil
}
