// Copyright (c) 2022 Gobalsky Labs Limited
//
// Use of this software is governed by the Business Source License included
// in the LICENSE file and at https://www.mariadb.com/bsl11.
//
// Change Date: 18 months from the later of the date of the first publicly
// available Distribution of this version of the repository, and 25 June 2022.
//
// On the date above, in accordance with the Business Source License, use
// of this software will be governed by version 3 or later of the GNU General
// Public License.

package gql

import (
	"context"
	"errors"

	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/data-node/vegatime"
	protoapi "code.vegaprotocol.io/protos/data-node/api/v1"
	v2 "code.vegaprotocol.io/protos/data-node/api/v2"
	types "code.vegaprotocol.io/protos/vega"
)

type myMarketResolver VegaResolverRoot

func (r *myMarketResolver) LiquidityProvisions(
	ctx context.Context,
	market *types.Market,
	party *string,
) ([]*types.LiquidityProvision, error) {
	var pid string
	if party != nil {
		pid = *party
	}

	req := protoapi.LiquidityProvisionsRequest{
		Party:  pid,
		Market: market.Id,
	}
	res, err := r.tradingDataClient.LiquidityProvisions(ctx, &req)
	if err != nil {
		r.log.Error("tradingData client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}

	return res.LiquidityProvisions, nil
}

func (r *myMarketResolver) LiquidityProvisionsConnection(
	ctx context.Context,
	market *types.Market,
	party *string,
	pagination *v2.Pagination,
) (*v2.LiquidityProvisionsConnection, error) {
	var pid string
	if party != nil {
		pid = *party
	}

	var marketID string
	if market != nil {
		marketID = market.Id
	}

	req := v2.ListLiquidityProvisionsRequest{
		PartyId:    &pid,
		MarketId:   &marketID,
		Pagination: pagination,
	}

	res, err := r.tradingDataClientV2.ListLiquidityProvisions(ctx, &req)
	if err != nil {
		r.log.Error("tradingData client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}

	return res.LiquidityProvisions, nil
}

func (r *myMarketResolver) Data(ctx context.Context, market *types.Market) (*types.MarketData, error) {
	req := protoapi.MarketDataByIDRequest{
		MarketId: market.Id,
	}
	res, err := r.tradingDataClient.MarketDataByID(ctx, &req)
	if err != nil {
		r.log.Error("tradingData client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}
	return res.MarketData, nil
}

func (r *myMarketResolver) Orders(ctx context.Context, market *types.Market,
	skip, first, last *int,
) ([]*types.Order, error) {
	p := makePagination(skip, first, last)
	req := protoapi.OrdersByMarketRequest{
		MarketId:   market.Id,
		Pagination: p,
	}
	res, err := r.tradingDataClient.OrdersByMarket(ctx, &req)
	if err != nil {
		r.log.Error("tradingData client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}
	return res.Orders, nil
}

func (r *myMarketResolver) OrdersConnection(ctx context.Context, market *types.Market, pagination *v2.Pagination) (*v2.OrderConnection, error) {
	req := v2.ListOrdersRequest{
		MarketId:   &market.Id,
		Pagination: pagination,
	}

	res, err := r.tradingDataClientV2.ListOrders(ctx, &req)
	if err != nil {
		r.log.Error("tradingData client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}

	return res.Orders, nil
}

func (r *myMarketResolver) Trades(ctx context.Context, market *types.Market,
	skip, first, last *int,
) ([]*types.Trade, error) {
	p := makePagination(skip, first, last)
	req := protoapi.TradesByMarketRequest{
		MarketId:   market.Id,
		Pagination: p,
	}
	res, err := r.tradingDataClient.TradesByMarket(ctx, &req)
	if err != nil {
		r.log.Error("tradingData client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}

	return res.Trades, nil
}

func (r *myMarketResolver) TradesConnection(ctx context.Context, market *types.Market, pagination *v2.Pagination) (*v2.TradeConnection, error) {
	req := v2.ListTradesRequest{
		MarketId:   &market.Id,
		Pagination: pagination,
	}
	res, err := r.tradingDataClientV2.ListTrades(ctx, &req)
	if err != nil {
		r.log.Error("tradingData client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}
	return res.Trades, nil
}

func (r *myMarketResolver) Depth(ctx context.Context, market *types.Market, maxDepth *int) (*types.MarketDepth, error) {
	if market == nil {
		return nil, errors.New("market missing or empty")
	}

	req := protoapi.MarketDepthRequest{MarketId: market.Id}
	if maxDepth != nil {
		if *maxDepth <= 0 {
			return nil, errors.New("invalid maxDepth, must be a positive number")
		}
		req.MaxDepth = uint64(*maxDepth)
	}

	// Look for market depth for the given market (will validate market internally)
	// Note: Market depth is also known as OrderBook depth within the matching-engine
	res, err := r.tradingDataClient.MarketDepth(ctx, &req)
	if err != nil {
		r.log.Error("trading data client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}

	return &types.MarketDepth{
		MarketId:       res.MarketId,
		Buy:            res.Buy,
		Sell:           res.Sell,
		SequenceNumber: res.SequenceNumber,
	}, nil
}

func (r *myMarketResolver) Candles(ctx context.Context, market *types.Market,
	sinceRaw string, interval Interval,
) ([]*types.Candle, error) {
	pinterval, err := convertIntervalToProto(interval)
	if err != nil {
		r.log.Debug("interval convert error", logging.Error(err))
	}

	since, err := vegatime.Parse(sinceRaw)
	if err != nil {
		return nil, err
	}

	var mkt string
	if market != nil {
		mkt = market.Id
	}

	req := protoapi.CandlesRequest{
		MarketId:       mkt,
		SinceTimestamp: since.UnixNano(),
		Interval:       pinterval,
	}
	res, err := r.tradingDataClient.Candles(ctx, &req)
	if err != nil {
		r.log.Error("tradingData client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}
	return res.Candles, nil
}

// Accounts ...
// if partyID specified get margin account for the given market
// if nil return the insurance pool & Fee Liquidty accounts for the market
func (r *myMarketResolver) Accounts(ctx context.Context, market *types.Market, partyID *string) ([]*types.Account, error) {
	filter := v2.AccountFilter{MarketIds: []string{market.Id}}
	ptyID := ""

	if partyID != nil {
		// get margin account for a party
		ptyID = *partyID
		filter.PartyIds = []string{ptyID}
		filter.AccountTypes = []types.AccountType{types.AccountType_ACCOUNT_TYPE_MARGIN}
	} else {
		filter.AccountTypes = []types.AccountType{
			types.AccountType_ACCOUNT_TYPE_INSURANCE,
			types.AccountType_ACCOUNT_TYPE_FEES_LIQUIDITY,
		}
	}

	req := v2.ListAccountsRequest{Filter: &filter}

	res, err := r.tradingDataClientV2.ListAccounts(ctx, &req)
	if err != nil {
		r.log.Error("unable to get market accounts",
			logging.Error(err),
			logging.String("market-id", market.Id),
			logging.String("party-id", ptyID))
		return []*types.Account{}, customErrorFromStatus(err)
	}

	if len(res.Accounts.Edges) == 0 {
		return []*types.Account{}, nil
	}

	accounts := make([]*types.Account, len(res.Accounts.Edges))
	for i, edge := range res.Accounts.Edges {
		accounts[i] = edge.Account
	}
	return accounts, nil
}

func (r *myMarketResolver) DecimalPlaces(ctx context.Context, obj *types.Market) (int, error) {
	return int(obj.DecimalPlaces), nil
}

func (r *myMarketResolver) PositionDecimalPlaces(ctx context.Context, obj *types.Market) (int, error) {
	return int(obj.PositionDecimalPlaces), nil
}

func (r *myMarketResolver) Name(ctx context.Context, obj *types.Market) (string, error) {
	return obj.TradableInstrument.Instrument.Name, nil
}

func (r *myMarketResolver) OpeningAuction(ctx context.Context, obj *types.Market) (*AuctionDuration, error) {
	return &AuctionDuration{
		DurationSecs: int(obj.OpeningAuction.Duration),
		Volume:       int(obj.OpeningAuction.Volume),
	}, nil
}

func (r *myMarketResolver) PriceMonitoringSettings(ctx context.Context, obj *types.Market) (*PriceMonitoringSettings, error) {
	return PriceMonitoringSettingsFromProto(obj.PriceMonitoringSettings)
}

func (r *myMarketResolver) LiquidityMonitoringParameters(ctx context.Context, obj *types.Market) (*LiquidityMonitoringParameters, error) {
	return &LiquidityMonitoringParameters{
		TargetStakeParameters: &TargetStakeParameters{
			TimeWindow:    int(obj.LiquidityMonitoringParameters.TargetStakeParameters.TimeWindow),
			ScalingFactor: obj.LiquidityMonitoringParameters.TargetStakeParameters.ScalingFactor,
		},
		TriggeringRatio: obj.LiquidityMonitoringParameters.TriggeringRatio,
	}, nil
}

func (r *myMarketResolver) TradingMode(ctx context.Context, obj *types.Market) (MarketTradingMode, error) {
	return convertMarketTradingModeFromProto(obj.TradingMode)
}

func (r *myMarketResolver) State(ctx context.Context, obj *types.Market) (MarketState, error) {
	return convertMarketStateFromProto(obj.State)
}

func (r *myMarketResolver) Proposal(ctx context.Context, obj *types.Market) (*types.GovernanceData, error) {
	resp, err := r.tradingDataClient.GetProposalByID(ctx, &protoapi.GetProposalByIDRequest{
		ProposalId: obj.Id,
	})
	// it's possible to not find a proposal as of now.
	// some market are loaded at startup, without
	// going through the proposal phase
	if err != nil {
		return nil, nil
	}
	return resp.Data, nil
}

func (r *myMarketResolver) RiskFactors(ctx context.Context, obj *types.Market) (*types.RiskFactor, error) {
	rf, err := r.tradingDataClient.GetRiskFactors(ctx, &protoapi.GetRiskFactorsRequest{
		MarketId: obj.Id,
	})
	if err != nil {
		return nil, err
	}

	return rf.RiskFactor, nil
}

func (r *myMarketResolver) CandlesConnection(ctx context.Context, market *types.Market, sinceRaw string, toRaw *string,
	interval Interval, pagination *v2.Pagination) (*v2.CandleDataConnection, error) {
	return handleCandleConnectionRequest(ctx, r.tradingDataClientV2, market, sinceRaw, toRaw, interval, pagination)
}
