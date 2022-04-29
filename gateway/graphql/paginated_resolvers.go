package gql

import (
	"context"
	"errors"
	"fmt"

	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/data-node/vegatime"
	protoapi "code.vegaprotocol.io/protos/data-node/api/v1"
	v2 "code.vegaprotocol.io/protos/data-node/api/v2"
	types "code.vegaprotocol.io/protos/vega"
)

type myPaginatedMarketResolver VegaResolverRoot

func (r *myPaginatedMarketResolver) LiquidityProvisions(
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

func (r *myPaginatedMarketResolver) Data(ctx context.Context, market *types.Market) (*types.MarketData, error) {
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

func (r *myPaginatedMarketResolver) Orders(ctx context.Context, market *types.Market,
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

func (r *myPaginatedMarketResolver) Trades(ctx context.Context, market *types.Market,
	first *int, after *string, last *int, before *string) (*v2.TradeConnection, error,
) {
	var firstLimit, lastLimit *int32

	if first != nil {
		v := int32(*first)
		firstLimit = &v
	}

	if last != nil {
		v := int32(*last)
		lastLimit = &v
	}

	req := v2.TradesByMarketRequest{
		MarketId: market.Id,
		Cursor: &v2.Cursor{
			First:  firstLimit,
			After:  after,
			Last:   lastLimit,
			Before: before,
		},
	}
	res, err := r.tradingDataClientV2.TradesByMarket(ctx, &req)
	if err != nil {
		r.log.Error("tradingData client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}
	return res.Trades, nil
}

func (r *myPaginatedMarketResolver) Depth(ctx context.Context, market *types.Market, maxDepth *int) (*types.MarketDepth, error) {
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

func (r *myPaginatedMarketResolver) Candles(ctx context.Context, market *types.Market,
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
// if nil return the insurance pool for the market
func (r *myPaginatedMarketResolver) Accounts(ctx context.Context, market *types.Market, partyID *string) ([]*types.Account, error) {
	// get margin account for a party
	if partyID != nil {
		req := protoapi.PartyAccountsRequest{
			PartyId:  *partyID,
			MarketId: market.Id,
			Type:     types.AccountType_ACCOUNT_TYPE_MARGIN,
			Asset:    "",
		}
		res, err := r.tradingDataClient.PartyAccounts(ctx, &req)
		if err != nil {
			r.log.Error("unable to get PartyAccounts",
				logging.Error(err),
				logging.String("market-id", market.Id),
				logging.String("party-id", *partyID))
			return []*types.Account{}, customErrorFromStatus(err)
		}
		return res.Accounts, nil
	}
	// get accounts for the market
	req := protoapi.MarketAccountsRequest{
		MarketId: market.Id,
		Asset:    "", // all assets
	}
	res, err := r.tradingDataClient.MarketAccounts(ctx, &req)
	if err != nil {
		r.log.Error("unable to get MarketAccounts",
			logging.Error(err),
			logging.String("market-id", market.Id))
		return []*types.Account{}, customErrorFromStatus(err)
	}
	return res.Accounts, nil
}

func (r *myPaginatedMarketResolver) DecimalPlaces(ctx context.Context, obj *types.Market) (int, error) {
	return int(obj.DecimalPlaces), nil
}

func (r *myPaginatedMarketResolver) PositionDecimalPlaces(ctx context.Context, obj *types.Market) (int, error) {
	return int(obj.PositionDecimalPlaces), nil
}

func (r *myPaginatedMarketResolver) Name(ctx context.Context, obj *types.Market) (string, error) {
	return obj.TradableInstrument.Instrument.Name, nil
}

func (r *myPaginatedMarketResolver) OpeningAuction(ctx context.Context, obj *types.Market) (*AuctionDuration, error) {
	return &AuctionDuration{
		DurationSecs: int(obj.OpeningAuction.Duration),
		Volume:       int(obj.OpeningAuction.Volume),
	}, nil
}

func (r *myPaginatedMarketResolver) PriceMonitoringSettings(ctx context.Context, obj *types.Market) (*PriceMonitoringSettings, error) {
	return PriceMonitoringSettingsFromProto(obj.PriceMonitoringSettings)
}

func (r *myPaginatedMarketResolver) LiquidityMonitoringParameters(ctx context.Context, obj *types.Market) (*LiquidityMonitoringParameters, error) {
	return &LiquidityMonitoringParameters{
		TargetStakeParameters: &TargetStakeParameters{
			TimeWindow:    int(obj.LiquidityMonitoringParameters.TargetStakeParameters.TimeWindow),
			ScalingFactor: obj.LiquidityMonitoringParameters.TargetStakeParameters.ScalingFactor,
		},
		TriggeringRatio: obj.LiquidityMonitoringParameters.TriggeringRatio,
	}, nil
}

func (r *myPaginatedMarketResolver) TradingMode(ctx context.Context, obj *types.Market) (MarketTradingMode, error) {
	return convertMarketTradingModeFromProto(obj.TradingMode)
}

func (r *myPaginatedMarketResolver) State(ctx context.Context, obj *types.Market) (MarketState, error) {
	return convertMarketStateFromProto(obj.State)
}

func (r *myPaginatedMarketResolver) Proposal(ctx context.Context, obj *types.Market) (*types.GovernanceData, error) {
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

func (r *myPaginatedMarketResolver) RiskFactors(ctx context.Context, obj *types.Market) (*types.RiskFactor, error) {
	rf, err := r.tradingDataClient.GetRiskFactors(ctx, &protoapi.GetRiskFactorsRequest{
		MarketId: obj.Id,
	})
	if err != nil {
		return nil, err
	}

	return rf.RiskFactor, nil
}

type myPaginatedPartyResolver VegaResolverRoot

func (r *myPaginatedPartyResolver) Rewards(
	ctx context.Context,
	party *types.Party,
	asset *string,
	skip, first, last *int,
) ([]*types.Reward, error) {
	var assetID string
	if asset != nil {
		assetID = *asset
	}

	p := makePagination(skip, first, last)

	req := &protoapi.GetRewardsRequest{
		PartyId:    party.Id,
		AssetId:    assetID,
		Pagination: p,
	}
	resp, err := r.tradingDataClient.GetRewards(ctx, req)
	return resp.Rewards, err
}

func (r *myPaginatedPartyResolver) RewardSummaries(
	ctx context.Context,
	party *types.Party,
	asset *string) ([]*types.RewardSummary, error,
) {
	var assetID string
	if asset != nil {
		assetID = *asset
	}

	req := &protoapi.GetRewardSummariesRequest{
		PartyId: party.Id,
		AssetId: assetID,
	}

	resp, err := r.tradingDataClient.GetRewardSummaries(ctx, req)
	return resp.Summaries, err
}

func (r *myPaginatedPartyResolver) Stake(
	ctx context.Context,
	party *types.Party,
) (*protoapi.PartyStakeResponse, error) {
	return r.tradingDataClient.PartyStake(
		ctx, &protoapi.PartyStakeRequest{
			Party: party.Id,
		},
	)
}

func (r *myPaginatedPartyResolver) LiquidityProvisions(
	ctx context.Context,
	party *types.Party,
	market, ref *string,
) ([]*types.LiquidityProvision, error) {
	var mid string
	if market != nil {
		mid = *market
	}

	req := protoapi.LiquidityProvisionsRequest{
		Party:  party.Id,
		Market: mid,
	}
	res, err := r.tradingDataClient.LiquidityProvisions(ctx, &req)
	if err != nil {
		r.log.Error("tradingData client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}

	var out []*types.LiquidityProvision
	if ref != nil {
		for _, v := range res.LiquidityProvisions {
			if v.Reference == *ref {
				out = append(out, v)
			}
		}
	} else {
		out = res.LiquidityProvisions
	}

	return out, nil
}

func (r *myPaginatedPartyResolver) Margins(ctx context.Context,
	party *types.Party, marketID *string) ([]*types.MarginLevels, error,
) {
	req := protoapi.MarginLevelsRequest{
		PartyId: party.Id,
	}
	if marketID != nil {
		req.MarketId = *marketID
	}

	res, err := r.tradingDataClient.MarginLevels(ctx, &req)
	if err != nil {
		r.log.Error("tradingData client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}
	out := make([]*types.MarginLevels, 0, len(res.MarginLevels))
	out = append(out, res.MarginLevels...)
	return out, nil
}

func (r *myPaginatedPartyResolver) Orders(ctx context.Context, party *types.Party,
	skip, first, last *int) ([]*types.Order, error,
) {
	p := makePagination(skip, first, last)
	req := protoapi.OrdersByPartyRequest{
		PartyId:    party.Id,
		Pagination: p,
	}
	res, err := r.tradingDataClient.OrdersByParty(ctx, &req)
	if err != nil {
		r.log.Error("tradingData client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}

	if len(res.Orders) > 0 {
		return res.Orders, nil
	}
	// mandatory return field in schema
	return []*types.Order{}, nil
}

func (r *myPaginatedPartyResolver) Trades(ctx context.Context, party *types.Party, market *string,
	first *int, after *string, last *int, before *string,
) (*v2.TradeConnection, error) {
	var mkt string
	if market != nil {
		mkt = *market
	}

	var firstLimit, lastLimit *int32

	if first != nil {
		v := int32(*first)
		firstLimit = &v
	}

	if last != nil {
		v := int32(*last)
		lastLimit = &v
	}

	req := v2.TradesByPartyRequest{
		PartyId:  party.Id,
		MarketId: mkt,
		Cursor: &v2.Cursor{
			First:  firstLimit,
			After:  after,
			Last:   lastLimit,
			Before: before,
		},
	}

	res, err := r.tradingDataClientV2.TradesByParty(ctx, &req)
	if err != nil {
		r.log.Error("tradingData client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}
	return res.Trades, nil
}

func (r *myPaginatedPartyResolver) Positions(ctx context.Context, party *types.Party) ([]*types.Position, error) {
	if party == nil {
		return nil, errors.New("nil party")
	}
	req := protoapi.PositionsByPartyRequest{PartyId: party.Id}
	res, err := r.tradingDataClient.PositionsByParty(ctx, &req)
	if err != nil {
		r.log.Error("tradingData client", logging.Error(err))
		return nil, customErrorFromStatus(err)
	}
	if len(res.Positions) > 0 {
		return res.Positions, nil
	}
	// mandatory return field in schema
	return []*types.Position{}, nil
}

func (r *myPaginatedPartyResolver) Accounts(ctx context.Context, party *types.Party,
	marketID *string, asset *string, accType *types.AccountType,
) ([]*types.Account, error) {
	if party == nil {
		return nil, errors.New("a party must be specified when querying accounts")
	}
	var (
		mktid = ""
		asst  = ""
		accTy = types.AccountType_ACCOUNT_TYPE_UNSPECIFIED
		err   error
	)

	if marketID != nil {
		mktid = *marketID
	}
	if asset != nil {
		asst = *asset
	}
	if accType != nil {
		accTy = *accType
		if err != nil ||
			(accTy != types.AccountType_ACCOUNT_TYPE_GENERAL &&
				accTy != types.AccountType_ACCOUNT_TYPE_MARGIN &&
				accTy != types.AccountType_ACCOUNT_TYPE_LOCK_WITHDRAW &&
				accTy != types.AccountType_ACCOUNT_TYPE_BOND) {
			return nil, fmt.Errorf("invalid account type for party %v", accType)
		}
	}
	req := protoapi.PartyAccountsRequest{
		PartyId:  party.Id,
		MarketId: mktid,
		Asset:    asst,
		Type:     accTy,
	}
	res, err := r.tradingDataClient.PartyAccounts(ctx, &req)
	if err != nil {
		r.log.Error("unable to get Party account",
			logging.Error(err),
			logging.String("party-id", party.Id),
			logging.String("market-id", mktid),
			logging.String("asset", asst),
			logging.String("type", accTy.String()))
		return nil, customErrorFromStatus(err)
	}

	if len(res.Accounts) > 0 {
		return res.Accounts, nil
	}
	// mandatory return field in schema
	return []*types.Account{}, nil
}

func (r *myPaginatedPartyResolver) Proposals(ctx context.Context, party *types.Party, inState *ProposalState) ([]*types.GovernanceData, error) {
	filter, err := inState.ToOptionalProposalState()
	if err != nil {
		return nil, err
	}
	resp, err := r.tradingDataClient.GetProposalsByParty(ctx, &protoapi.GetProposalsByPartyRequest{
		PartyId:       party.Id,
		SelectInState: filter,
	})
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (r *myPaginatedPartyResolver) Withdrawals(ctx context.Context, party *types.Party) ([]*types.Withdrawal, error) {
	res, err := r.tradingDataClient.Withdrawals(
		ctx, &protoapi.WithdrawalsRequest{PartyId: party.Id},
	)
	if err != nil {
		return nil, err
	}

	return res.Withdrawals, nil
}

func (r *myPaginatedPartyResolver) Deposits(ctx context.Context, party *types.Party) ([]*types.Deposit, error) {
	res, err := r.tradingDataClient.Deposits(
		ctx, &protoapi.DepositsRequest{PartyId: party.Id},
	)
	if err != nil {
		return nil, err
	}

	return res.Deposits, nil
}

func (r *myPaginatedPartyResolver) Votes(ctx context.Context, party *types.Party) ([]*ProposalVote, error) {
	resp, err := r.tradingDataClient.GetVotesByParty(ctx, &protoapi.GetVotesByPartyRequest{
		PartyId: party.Id,
	})
	if err != nil {
		return nil, err
	}
	result := make([]*ProposalVote, len(resp.Votes))
	for i, vote := range resp.Votes {
		result[i] = ProposalVoteFromProto(vote)
	}
	return result, nil
}

func (r *myPaginatedPartyResolver) Delegations(
	ctx context.Context,
	obj *types.Party,
	nodeID *string,
	skip, first, last *int,
) ([]*types.Delegation, error) {
	req := &protoapi.DelegationsRequest{
		Party:      obj.Id,
		Pagination: makePagination(skip, first, last),
	}

	if nodeID != nil {
		req.NodeId = *nodeID
	}

	resp, err := r.tradingDataClient.Delegations(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Delegations, nil
}

// END: Party Resolver
