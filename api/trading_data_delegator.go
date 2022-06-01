package api

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"code.vegaprotocol.io/data-node/candlesv2"
	"code.vegaprotocol.io/data-node/risk"
	"code.vegaprotocol.io/data-node/vegatime"
	commandspb "code.vegaprotocol.io/protos/vega/commands/v1"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vega/types/num"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/metrics"
	"code.vegaprotocol.io/data-node/sqlstore"
	protoapi "code.vegaprotocol.io/protos/data-node/api/v1"
	"code.vegaprotocol.io/protos/vega"
	oraclespb "code.vegaprotocol.io/protos/vega/oracles/v1"
	"google.golang.org/grpc/codes"
)

type tradingDataDelegator struct {
	*tradingDataService
	orderStore              *sqlstore.Orders
	tradeStore              *sqlstore.Trades
	assetStore              *sqlstore.Assets
	accountStore            *sqlstore.Accounts
	marketDataStore         *sqlstore.MarketData
	rewardStore             *sqlstore.Rewards
	marketsStore            *sqlstore.Markets
	delegationStore         *sqlstore.Delegations
	epochStore              *sqlstore.Epochs
	depositsStore           *sqlstore.Deposits
	withdrawalsStore        *sqlstore.Withdrawals
	proposalsStore          *sqlstore.Proposals
	voteStore               *sqlstore.Votes
	riskFactorStore         *sqlstore.RiskFactors
	marginLevelsStore       *sqlstore.MarginLevels
	netParamStore           *sqlstore.NetworkParameters
	blockStore              *sqlstore.Blocks
	checkpointStore         *sqlstore.Checkpoints
	partyStore              *sqlstore.Parties
	candleServiceV2         *candlesv2.Svc
	oracleSpecStore         *sqlstore.OracleSpec
	oracleDataStore         *sqlstore.OracleData
	liquidityProvisionStore *sqlstore.LiquidityProvision
	positionStore           *sqlstore.Positions
	transfersStore          *sqlstore.Transfers
	stakingStore            *sqlstore.StakeLinking
	notaryStore             *sqlstore.Notary
	keyRotationsStore       *sqlstore.KeyRotations
	nodeStore               *sqlstore.Node
}

var defaultEntityPagination = entities.OffsetPagination{
	Skip:       0,
	Limit:      50,
	Descending: true,
}

/****************************** Positions **************************************/
func (t *tradingDataDelegator) PositionsByParty(ctx context.Context, request *protoapi.PositionsByPartyRequest) (*protoapi.PositionsByPartyResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("PositionsByParty SQL")()

	p := defaultPaginationV2
	if request.Pagination != nil {
		p = toEntityPagination(request.Pagination)
	}

	var positions []entities.Position
	var err error

	if request.MarketId == "" && request.PartyId == "" {
		positions, err = t.positionStore.GetAll(ctx)
	} else if request.MarketId == "" {
		positions, err = t.positionStore.GetByParty(ctx, entities.NewPartyID(request.PartyId), &p)
	} else if request.PartyId == "" {
		positions, err = t.positionStore.GetByMarket(ctx, entities.NewMarketID(request.MarketId), &p)
	} else {
		positions = make([]entities.Position, 1)
		positions[0], err = t.positionStore.GetByMarketAndParty(ctx,
			entities.NewMarketID(request.MarketId),
			entities.NewPartyID(request.PartyId))

		// Don't error if there's no position for this party/market
		if errors.Is(err, sqlstore.ErrPositionNotFound) {
			err = nil
			positions = []entities.Position{}
		}
	}

	if err != nil {
		return nil, apiError(codes.Internal, ErrTradeServiceGetPositionsByParty, err)
	}

	out := make([]*vega.Position, len(positions))
	for i, position := range positions {
		out[i] = position.ToProto()
	}

	response := &protoapi.PositionsByPartyResponse{Positions: out}
	return response, nil
}

/****************************** Parties **************************************/
func (t *tradingDataDelegator) Parties(ctx context.Context, _ *protoapi.PartiesRequest) (*protoapi.PartiesResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("Parties SQL")()
	parties, err := t.partyStore.GetAll(ctx)
	if err != nil {
		return nil, apiError(codes.Internal, ErrPartyServiceGetAll, err)
	}

	out := make([]*vega.Party, len(parties))
	for i, p := range parties {
		out[i] = p.ToProto()
	}

	return &protoapi.PartiesResponse{
		Parties: out,
	}, nil
}

func (t *tradingDataDelegator) PartyByID(ctx context.Context, req *protoapi.PartyByIDRequest) (*protoapi.PartyByIDResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("PartyByID SQL")()
	out := protoapi.PartyByIDResponse{}

	party, err := t.partyStore.GetByID(ctx, req.PartyId)

	if errors.Is(err, sqlstore.ErrPartyNotFound) {
		return &out, nil
	}

	if errors.Is(err, sqlstore.ErrInvalidPartyID) {
		return &out, apiError(codes.InvalidArgument, ErrPartyServiceGetByID, err)
	}

	if err != nil {
		return nil, apiError(codes.Internal, ErrPartyServiceGetByID, err)
	}

	return &protoapi.PartyByIDResponse{
		Party: party.ToProto(),
	}, nil
}

/****************************** General **************************************/

func (t *tradingDataDelegator) GetVegaTime(ctx context.Context, _ *protoapi.GetVegaTimeRequest) (*protoapi.GetVegaTimeResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetVegaTime SQL")()
	b, err := t.blockStore.GetLastBlock(ctx)
	if err != nil {
		return nil, apiError(codes.Internal, ErrTimeServiceGetTimeNow, err)
	}

	return &protoapi.GetVegaTimeResponse{
		Timestamp: b.VegaTime.UnixNano(),
	}, nil
}

/****************************** Checkpoints **************************************/

func (t *tradingDataDelegator) Checkpoints(ctx context.Context, _ *protoapi.CheckpointsRequest) (*protoapi.CheckpointsResponse, error) {
	checkpoints, err := t.checkpointStore.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*protoapi.Checkpoint, len(checkpoints))
	for i, cp := range checkpoints {
		out[i] = cp.ToProto()
	}

	return &protoapi.CheckpointsResponse{
		Checkpoints: out,
	}, nil
}

/****************************** Transfers **************************************/

func (t *tradingDataDelegator) Transfers(ctx context.Context, req *protoapi.TransfersRequest) (*protoapi.TransfersResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("Transfers-SQL")()

	if len(req.Pubkey) <= 0 && (req.IsFrom || req.IsTo) {
		return nil, apiError(codes.InvalidArgument, errors.New("missing pubkey"))
	}

	if req.IsFrom && req.IsTo {
		return nil, apiError(codes.InvalidArgument, errors.New("request is for transfers to and from the same party"))
	}

	var transfers []*entities.Transfer
	var err error
	if !req.IsFrom && !req.IsTo {
		transfers, err = t.transfersStore.GetAll(ctx)
		if err != nil {
			return nil, apiError(codes.Internal, err)
		}
	} else if req.IsFrom || req.IsTo {

		if req.IsFrom {
			transfers, err = t.transfersStore.GetTransfersFromParty(ctx, entities.PartyID{ID: entities.ID(req.Pubkey)})
			if err != nil {
				return nil, apiError(codes.Internal, err)
			}
		}

		if req.IsTo {
			transfers, err = t.transfersStore.GetTransfersToParty(ctx, entities.PartyID{ID: entities.ID(req.Pubkey)})
			if err != nil {
				return nil, apiError(codes.Internal, err)
			}
		}
	}

	protoTransfers := make([]*eventspb.Transfer, 0, len(transfers))
	for _, transfer := range transfers {
		proto, err := transfer.ToProto(t.accountStore)
		if err != nil {
			return nil, apiError(codes.Internal, err)
		}
		protoTransfers = append(protoTransfers, proto)
	}

	return &protoapi.TransfersResponse{
		Transfers: protoTransfers,
	}, nil
}

/****************************** Network Parameters **************************************/

func (t *tradingDataDelegator) NetworkParameters(ctx context.Context, req *protoapi.NetworkParametersRequest) (*protoapi.NetworkParametersResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("NetworkParameters SQL")()
	nps, err := t.netParamStore.GetAll(ctx)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	out := make([]*vega.NetworkParameter, len(nps))
	for i, np := range nps {
		out[i] = np.ToProto()
	}

	return &protoapi.NetworkParametersResponse{
		NetworkParameters: out,
	}, nil
}

/****************************** Candles **************************************/

func (t *tradingDataDelegator) Candles(ctx context.Context,
	request *protoapi.CandlesRequest,
) (*protoapi.CandlesResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("Candles-SQL")()

	if request.Interval == vega.Interval_INTERVAL_UNSPECIFIED {
		return nil, apiError(codes.InvalidArgument, ErrMalformedRequest)
	}

	from := vegatime.UnixNano(request.SinceTimestamp)
	interval, err := toV2IntervalString(request.Interval)
	if err != nil {
		return nil, apiError(codes.Internal, ErrCandleServiceGetCandleData,
			fmt.Errorf("failed to get candles:%w", err))
	}

	exists, candleId, err := t.candleServiceV2.GetCandleIdForIntervalAndMarket(ctx, interval, request.MarketId)
	if err != nil {
		return nil, apiError(codes.Internal, ErrCandleServiceGetCandleData,
			fmt.Errorf("failed to get candles:%w", err))
	}

	if !exists {
		return nil, apiError(codes.InvalidArgument, ErrCandleServiceGetCandleData,
			fmt.Errorf("candle does not exist for interval %s and market %s", interval, request.MarketId))
	}

	candles, err := t.candleServiceV2.GetCandleDataForTimeSpan(ctx, candleId, &from, nil, entities.OffsetPagination{})
	if err != nil {
		return nil, apiError(codes.Internal, ErrCandleServiceGetCandleData,
			fmt.Errorf("failed to get candles for interval:%w", err))
	}

	var protoCandles []*vega.Candle
	for _, candle := range candles {
		proto, err := candle.ToV1CandleProto(request.Interval)
		if err != nil {
			return nil, apiError(codes.Internal, ErrCandleServiceGetCandleData,
				fmt.Errorf("failed to convert candle to protobuf:%w", err))
		}

		protoCandles = append(protoCandles, proto)
	}

	return &protoapi.CandlesResponse{
		Candles: protoCandles,
	}, nil
}

func toV2IntervalString(interval vega.Interval) (string, error) {
	switch interval {
	case vega.Interval_INTERVAL_I1M:
		return "1 minute", nil
	case vega.Interval_INTERVAL_I5M:
		return "5 minutes", nil
	case vega.Interval_INTERVAL_I15M:
		return "15 minutes", nil
	case vega.Interval_INTERVAL_I1H:
		return "1 hour", nil
	case vega.Interval_INTERVAL_I6H:
		return "6 hours", nil
	case vega.Interval_INTERVAL_I1D:
		return "1 day", nil
	default:
		return "", fmt.Errorf("interval not support:%s", interval)
	}
}

func (t *tradingDataDelegator) CandlesSubscribe(req *protoapi.CandlesSubscribeRequest,
	srv protoapi.TradingDataService_CandlesSubscribeServer,
) error {
	defer metrics.StartAPIRequestAndTimeGRPC("CandlesSubscribe-SQL")()
	// Wrap context from the request into cancellable. We can close internal chan on error.
	ctx, cancel := context.WithCancel(srv.Context())
	defer cancel()

	interval, err := toV2IntervalString(req.Interval)
	if err != nil {
		return apiError(codes.InvalidArgument, ErrStreamInternal,
			fmt.Errorf("subscribing to candles:%w", err))
	}

	exists, candleId, err := t.candleServiceV2.GetCandleIdForIntervalAndMarket(ctx, interval, req.MarketId)
	if err != nil {
		return apiError(codes.InvalidArgument, ErrStreamInternal,
			fmt.Errorf("subscribing to candles:%w", err))
	}

	if !exists {
		return apiError(codes.InvalidArgument, ErrStreamInternal,
			fmt.Errorf("candle does not exist for interval %s and market %s", interval, req.MarketId))
	}

	ref, candlesChan, err := t.candleServiceV2.Subscribe(ctx, candleId)
	if err != nil {
		return apiError(codes.Internal, ErrStreamInternal,
			fmt.Errorf("subscribing to candles:%w", err))
	}

	for {
		select {
		case candle, ok := <-candlesChan:

			if !ok {
				err = ErrChannelClosed
				return apiError(codes.Internal, err)
			}
			proto, err := candle.ToV1CandleProto(req.Interval)
			if err != nil {
				return apiError(codes.Internal, ErrStreamInternal, err)
			}

			resp := &protoapi.CandlesSubscribeResponse{
				Candle: proto,
			}
			if err = srv.Send(resp); err != nil {
				return apiError(codes.Internal, ErrStreamInternal, err)
			}
		case <-ctx.Done():
			err := t.candleServiceV2.Unsubscribe(ref)
			if err != nil {
				t.log.Errorf("failed to unsubscribe from candle updates:%s", err)
			}
			return apiError(codes.Internal, ErrStreamInternal, ctx.Err())
		}
	}
}

/****************************** Governance **************************************/

func (t *tradingDataDelegator) GetProposals(ctx context.Context, req *protoapi.GetProposalsRequest,
) (*protoapi.GetProposalsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetProposals SQL")()

	inState := proposalState(req.SelectInState)

	proposals, err := t.proposalsStore.Get(ctx, inState, nil, nil)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	governanceData, err := t.proposalListToGovernanceData(ctx, proposals)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	return &protoapi.GetProposalsResponse{Data: governanceData}, nil
}

func (t *tradingDataDelegator) GetProposalsByParty(ctx context.Context,
	req *protoapi.GetProposalsByPartyRequest,
) (*protoapi.GetProposalsByPartyResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetProposalsByParty SQL")()

	inState := proposalState(req.SelectInState)

	proposals, err := t.proposalsStore.Get(ctx, inState, &req.PartyId, nil)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	governanceData, err := t.proposalListToGovernanceData(ctx, proposals)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	return &protoapi.GetProposalsByPartyResponse{
		Data: governanceData,
	}, nil
}

func (t *tradingDataDelegator) GetProposalByID(ctx context.Context,
	req *protoapi.GetProposalByIDRequest,
) (*protoapi.GetProposalByIDResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetProposalByID SQL")()

	proposal, err := t.proposalsStore.GetByID(ctx, req.ProposalId)
	if errors.Is(err, sqlstore.ErrProposalNotFound) {
		return nil, apiError(codes.NotFound, ErrMissingProposalID, err)
	} else if err != nil {
		return nil, apiError(codes.Internal, ErrNotMapped, err)
	}

	gd, err := t.proposalToGovernanceData(ctx, proposal)
	if err != nil {
		return nil, apiError(codes.Internal, ErrNotMapped, err)
	}

	return &protoapi.GetProposalByIDResponse{Data: gd}, nil
}

func (t *tradingDataDelegator) GetProposalByReference(ctx context.Context,
	req *protoapi.GetProposalByReferenceRequest,
) (*protoapi.GetProposalByReferenceResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetProposalByID SQL")()

	proposal, err := t.proposalsStore.GetByReference(ctx, req.Reference)
	if errors.Is(err, sqlstore.ErrProposalNotFound) {
		return nil, apiError(codes.NotFound, ErrMissingProposalReference, err)
	} else if err != nil {
		return nil, apiError(codes.Internal, ErrNotMapped, err)
	}

	gd, err := t.proposalToGovernanceData(ctx, proposal)
	if err != nil {
		return nil, apiError(codes.Internal, ErrNotMapped, err)
	}

	return &protoapi.GetProposalByReferenceResponse{Data: gd}, nil
}

func (t *tradingDataDelegator) GetVotesByParty(ctx context.Context,
	req *protoapi.GetVotesByPartyRequest,
) (*protoapi.GetVotesByPartyResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetVotesByParty SQL")()

	votes, err := t.voteStore.GetByParty(ctx, req.PartyId)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	return &protoapi.GetVotesByPartyResponse{Votes: voteListToProto(votes)}, nil
}

func (t *tradingDataDelegator) GetNewMarketProposals(ctx context.Context,
	req *protoapi.GetNewMarketProposalsRequest,
) (*protoapi.GetNewMarketProposalsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetNewMarketProposals SQL")()

	inState := proposalState(req.SelectInState)
	proposals, err := t.proposalsStore.Get(ctx, inState, nil, &entities.ProposalTypeNewMarket)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}
	governanceData, err := t.proposalListToGovernanceData(ctx, proposals)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}
	return &protoapi.GetNewMarketProposalsResponse{Data: governanceData}, nil
}

func (t *tradingDataDelegator) GetUpdateMarketProposals(ctx context.Context,
	req *protoapi.GetUpdateMarketProposalsRequest,
) (*protoapi.GetUpdateMarketProposalsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetUpdateMarketProposals SQL")()

	inState := proposalState(req.SelectInState)
	proposals, err := t.proposalsStore.Get(ctx, inState, nil, &entities.ProposalTypeUpdateMarket)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}
	governanceData, err := t.proposalListToGovernanceData(ctx, proposals)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}
	return &protoapi.GetUpdateMarketProposalsResponse{Data: governanceData}, nil
}

func (t *tradingDataDelegator) GetNetworkParametersProposals(ctx context.Context,
	req *protoapi.GetNetworkParametersProposalsRequest,
) (*protoapi.GetNetworkParametersProposalsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetNetworkParametersProposals SQL")()

	inState := proposalState(req.SelectInState)

	proposals, err := t.proposalsStore.Get(ctx, inState, nil, &entities.ProposalTypeUpdateNetworkParameter)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	governanceData, err := t.proposalListToGovernanceData(ctx, proposals)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	return &protoapi.GetNetworkParametersProposalsResponse{Data: governanceData}, nil
}

func (t *tradingDataDelegator) GetNewAssetProposals(ctx context.Context,
	req *protoapi.GetNewAssetProposalsRequest,
) (*protoapi.GetNewAssetProposalsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetNewAssetProposals SQL")()

	inState := proposalState(req.SelectInState)

	proposals, err := t.proposalsStore.Get(ctx, inState, nil, &entities.ProposalTypeNewAsset)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	governanceData, err := t.proposalListToGovernanceData(ctx, proposals)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	return &protoapi.GetNewAssetProposalsResponse{Data: governanceData}, nil
}

func (t *tradingDataDelegator) GetNewFreeformProposals(ctx context.Context,
	req *protoapi.GetNewFreeformProposalsRequest,
) (*protoapi.GetNewFreeformProposalsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetNewFreeformProposals SQL")()

	inState := proposalState(req.SelectInState)

	proposals, err := t.proposalsStore.Get(ctx, inState, nil, &entities.ProposalTypeNewFreeform)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	governanceData, err := t.proposalListToGovernanceData(ctx, proposals)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	return &protoapi.GetNewFreeformProposalsResponse{Data: governanceData}, nil
}

func proposalState(protoState *protoapi.OptionalProposalState) *entities.ProposalState {
	var s *entities.ProposalState
	if protoState != nil {
		state := entities.ProposalState(protoState.Value)
		s = &state
	}
	return s
}

func (t *tradingDataDelegator) proposalListToGovernanceData(ctx context.Context, proposals []entities.Proposal) ([]*vega.GovernanceData, error) {
	governanceData := make([]*vega.GovernanceData, len(proposals))
	for i, proposal := range proposals {
		gd, err := t.proposalToGovernanceData(ctx, proposal)
		if err != nil {
			return nil, err
		}
		governanceData[i] = gd
	}
	return governanceData, nil
}

func (t *tradingDataDelegator) proposalToGovernanceData(ctx context.Context, proposal entities.Proposal) (*vega.GovernanceData, error) {
	yesVotes, err := t.voteStore.GetYesVotesForProposal(ctx, proposal.ID.String())
	if err != nil {
		return nil, err
	}
	protoYesVotes := voteListToProto(yesVotes)

	noVotes, err := t.voteStore.GetNoVotesForProposal(ctx, proposal.ID.String())
	if err != nil {
		return nil, err
	}
	protoNoVotes := voteListToProto(noVotes)

	gd := vega.GovernanceData{
		Proposal: proposal.ToProto(),
		Yes:      protoYesVotes,
		No:       protoNoVotes,
	}
	return &gd, nil
}

func voteListToProto(votes []entities.Vote) []*vega.Vote {
	protoVotes := make([]*vega.Vote, len(votes))
	for j, vote := range votes {
		protoVotes[j] = vote.ToProto()
	}
	return protoVotes
}

/****************************** Epochs **************************************/

func (t *tradingDataDelegator) GetEpoch(ctx context.Context, req *protoapi.GetEpochRequest) (*protoapi.GetEpochResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetEpoch SQL")()

	var epoch entities.Epoch
	var err error

	if req.GetId() == 0 {
		epoch, err = t.epochStore.GetCurrent(ctx)
	} else {
		epoch, err = t.epochStore.Get(ctx, int64(req.GetId()))
	}

	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	protoEpoch := epoch.ToProto()

	delegations, err := t.delegationStore.Get(ctx, nil, nil, &epoch.ID, nil)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	protoDelegations := make([]*vega.Delegation, len(delegations))
	for i, delegation := range delegations {
		protoDelegations[i] = delegation.ToProto()
	}
	protoEpoch.Delegations = protoDelegations

	nodes, err := t.nodeStore.GetNodes(ctx, uint64(epoch.ID))
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	protoNodes := make([]*vega.Node, len(nodes))
	for i, node := range nodes {
		protoNodes[i] = node.ToProto()
	}

	protoEpoch.Validators = protoNodes

	return &protoapi.GetEpochResponse{
		Epoch: protoEpoch,
	}, nil
}

/****************************** Delegations **************************************/

func (t *tradingDataDelegator) Delegations(ctx context.Context,
	req *protoapi.DelegationsRequest,
) (*protoapi.DelegationsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("Delegations SQL")()

	var delegations []entities.Delegation
	var err error

	p := defaultPaginationV2
	if req.Pagination != nil {
		p = toEntityPagination(req.Pagination)
	}

	var epochID *int64
	var partyID *string
	var nodeID *string

	if req.EpochSeq != "" {
		epochNum, err := strconv.ParseInt(req.EpochSeq, 10, 64)
		if err != nil {
			return nil, apiError(codes.InvalidArgument, err)
		}
		epochID = &epochNum
	}

	if req.Party != "" {
		partyID = &req.Party
	}

	if req.NodeId != "" {
		nodeID = &req.NodeId
	}

	delegations, err = t.delegationStore.Get(ctx, partyID, nodeID, epochID, &p)

	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	protoDelegations := make([]*vega.Delegation, len(delegations))
	for i, delegation := range delegations {
		protoDelegations[i] = delegation.ToProto()
	}

	return &protoapi.DelegationsResponse{
		Delegations: protoDelegations,
	}, nil
}

/****************************** Rewards **************************************/

func (t *tradingDataDelegator) GetRewards(ctx context.Context,
	req *protoapi.GetRewardsRequest,
) (*protoapi.GetRewardsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetRewards-SQL")()
	if len(req.PartyId) <= 0 {
		return nil, apiError(codes.InvalidArgument, ErrGetRewards)
	}

	p := defaultPaginationV2
	if req.Pagination != nil {
		p = toEntityPagination(req.Pagination)
	}

	var rewards []entities.Reward
	var err error

	if len(req.AssetId) <= 0 {
		rewards, err = t.rewardStore.Get(ctx, &req.PartyId, nil, &p)
	} else {
		rewards, err = t.rewardStore.Get(ctx, &req.PartyId, &req.AssetId, &p)
	}

	if err != nil {
		return nil, apiError(codes.Internal, ErrGetRewards, err)
	}

	protoRewards := make([]*vega.Reward, len(rewards))
	for i, reward := range rewards {
		protoRewards[i] = reward.ToProto()
	}

	return &protoapi.GetRewardsResponse{Rewards: protoRewards}, nil
}

func (t *tradingDataDelegator) GetRewardSummaries(ctx context.Context,
	req *protoapi.GetRewardSummariesRequest,
) (*protoapi.GetRewardSummariesResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetRewardSummaries-SQL")()

	if len(req.PartyId) <= 0 {
		return nil, apiError(codes.InvalidArgument, ErrTradeServiceGetByParty)
	}

	var summaries []entities.RewardSummary
	var err error

	if len(req.AssetId) <= 0 {
		summaries, err = t.rewardStore.GetSummaries(ctx, &req.PartyId, nil)
	} else {
		summaries, err = t.rewardStore.GetSummaries(ctx, &req.PartyId, &req.AssetId)
	}

	if err != nil {
		return nil, apiError(codes.Internal, ErrGetRewards, err)
	}

	protoSummaries := make([]*vega.RewardSummary, len(summaries))
	for i, summary := range summaries {
		protoSummaries[i] = summary.ToProto()
	}

	return &protoapi.GetRewardSummariesResponse{Summaries: protoSummaries}, nil
}

/****************************** Trades **************************************/
// TradesByParty provides a list of trades for the given party.
// OffsetPagination: Optional. If not provided, defaults are used.
func (t *tradingDataDelegator) TradesByParty(ctx context.Context,
	req *protoapi.TradesByPartyRequest,
) (*protoapi.TradesByPartyResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("TradesByParty-SQL")()

	p := defaultEntityPagination
	if req.Pagination != nil {
		p = toEntityPagination(req.Pagination)
	}

	trades, err := t.tradeStore.GetByParty(ctx, req.PartyId, &req.MarketId, p)
	if err != nil {
		return nil, apiError(codes.Internal, ErrTradeServiceGetByParty, err)
	}

	protoTrades := tradesToProto(trades)

	return &protoapi.TradesByPartyResponse{Trades: protoTrades}, nil
}

func tradesToProto(trades []entities.Trade) []*vega.Trade {
	protoTrades := []*vega.Trade{}
	for _, trade := range trades {
		protoTrades = append(protoTrades, trade.ToProto())
	}
	return protoTrades
}

// TradesByOrder provides a list of the trades that correspond to a given order.
func (t *tradingDataDelegator) TradesByOrder(ctx context.Context,
	req *protoapi.TradesByOrderRequest,
) (*protoapi.TradesByOrderResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("TradesByOrder-SQL")()

	trades, err := t.tradeStore.GetByOrderID(ctx, req.OrderId, nil, defaultEntityPagination)
	if err != nil {
		return nil, apiError(codes.Internal, ErrTradeServiceGetByOrderID, err)
	}

	protoTrades := tradesToProto(trades)

	return &protoapi.TradesByOrderResponse{Trades: protoTrades}, nil
}

// TradesByMarket provides a list of trades for a given market.
// OffsetPagination: Optional. If not provided, defaults are used.
func (t *tradingDataDelegator) TradesByMarket(ctx context.Context, req *protoapi.TradesByMarketRequest) (*protoapi.TradesByMarketResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("TradesByMarket-SQL")()

	p := defaultEntityPagination
	if req.Pagination != nil {
		p = toEntityPagination(req.Pagination)
	}

	trades, err := t.tradeStore.GetByMarket(ctx, req.MarketId, p)
	if err != nil {
		return nil, apiError(codes.Internal, ErrTradeServiceGetByMarket, err)
	}

	protoTrades := tradesToProto(trades)
	return &protoapi.TradesByMarketResponse{
		Trades: protoTrades,
	}, nil
}

// LastTrade provides the last trade for the given market.
func (t *tradingDataDelegator) LastTrade(ctx context.Context,
	req *protoapi.LastTradeRequest,
) (*protoapi.LastTradeResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("LastTrade-SQL")()

	if len(req.MarketId) <= 0 {
		return nil, apiError(codes.InvalidArgument, ErrEmptyMissingMarketID)
	}

	p := entities.OffsetPagination{
		Skip:       0,
		Limit:      1,
		Descending: true,
	}

	trades, err := t.tradeStore.GetByMarket(ctx, req.MarketId, p)
	if err != nil {
		return nil, apiError(codes.Internal, ErrTradeServiceGetByMarket, err)
	}

	protoTrades := tradesToProto(trades)

	if len(protoTrades) > 0 && protoTrades[0] != nil {
		return &protoapi.LastTradeResponse{Trade: protoTrades[0]}, nil
	}
	// No trades found on the market yet (and no errors)
	// this can happen at the beginning of a new market
	return &protoapi.LastTradeResponse{}, nil
}

func (t *tradingDataDelegator) OrderByID(ctx context.Context, req *protoapi.OrderByIDRequest) (*protoapi.OrderByIDResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("OrderByID-SQL")()

	if len(req.OrderId) == 0 {
		return nil, ErrMissingOrderIDParameter
	}

	version := int32(req.Version)
	order, err := t.orderStore.GetByOrderID(ctx, req.OrderId, &version)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	resp := &protoapi.OrderByIDResponse{Order: order.ToProto()}
	return resp, nil
}

func (t *tradingDataDelegator) OrderByMarketAndID(ctx context.Context,
	req *protoapi.OrderByMarketAndIDRequest,
) (*protoapi.OrderByMarketAndIDResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("OrderByMarketAndID-SQL")()

	// This function is no longer needed; IDs are globally unique now, but keep it for compatibility for now
	if len(req.OrderId) == 0 {
		return nil, ErrMissingOrderIDParameter
	}

	order, err := t.orderStore.GetByOrderID(ctx, req.OrderId, nil)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	resp := &protoapi.OrderByMarketAndIDResponse{Order: order.ToProto()}
	return resp, nil
}

// OrderByReference provides the (possibly not yet accepted/rejected) order.
func (t *tradingDataDelegator) OrderByReference(ctx context.Context, req *protoapi.OrderByReferenceRequest) (*protoapi.OrderByReferenceResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("OrderByReference-SQL")()

	orders, err := t.orderStore.GetByReference(ctx, req.Reference, entities.OffsetPagination{})
	if err != nil {
		return nil, apiError(codes.InvalidArgument, ErrOrderServiceGetByReference, err)
	}

	if len(orders) == 0 {
		return nil, ErrOrderNotFound
	}
	return &protoapi.OrderByReferenceResponse{
		Order: orders[0].ToProto(),
	}, nil
}

func (t *tradingDataDelegator) OrdersByParty(ctx context.Context,
	req *protoapi.OrdersByPartyRequest,
) (*protoapi.OrdersByPartyResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("OrdersByParty-SQL")()

	p := defaultPaginationV2
	if req.Pagination != nil {
		p = toEntityPagination(req.Pagination)
	}

	orders, err := t.orderStore.GetByParty(ctx, req.PartyId, p)
	if err != nil {
		return nil, apiError(codes.InvalidArgument, ErrOrderServiceGetByParty, err)
	}

	pbOrders := make([]*vega.Order, len(orders))
	for i, order := range orders {
		pbOrders[i] = order.ToProto()
	}

	return &protoapi.OrdersByPartyResponse{
		Orders: pbOrders,
	}, nil
}

func toEntityPagination(pagination *protoapi.Pagination) entities.OffsetPagination {
	return entities.OffsetPagination{
		Skip:       pagination.Skip,
		Limit:      pagination.Limit,
		Descending: pagination.Descending,
	}
}

func (t *tradingDataDelegator) AssetByID(ctx context.Context, req *protoapi.AssetByIDRequest) (*protoapi.AssetByIDResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("AssetByID-SQL")()
	if len(req.Id) <= 0 {
		return nil, apiError(codes.InvalidArgument, errors.New("missing ID"))
	}

	asset, err := t.assetStore.GetByID(ctx, req.Id)
	if err != nil {
		return nil, apiError(codes.NotFound, err)
	}

	return &protoapi.AssetByIDResponse{
		Asset: asset.ToProto(),
	}, nil
}

func (t *tradingDataDelegator) Assets(ctx context.Context, _ *protoapi.AssetsRequest) (*protoapi.AssetsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("Assets-SQL")()

	assets, _ := t.assetStore.GetAll(ctx)

	out := make([]*vega.Asset, 0, len(assets))
	for _, v := range assets {
		out = append(out, v.ToProto())
	}
	return &protoapi.AssetsResponse{
		Assets: out,
	}, nil
}

func isValidAccountType(accountType vega.AccountType, validAccountTypes ...vega.AccountType) bool {
	for _, vt := range validAccountTypes {
		if accountType == vt {
			return true
		}
	}

	return false
}

func (t *tradingDataDelegator) PartyAccounts(ctx context.Context, req *protoapi.PartyAccountsRequest) (*protoapi.PartyAccountsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("PartyAccounts_SQL")()

	// This is just nicer to read and update if the list of valid account types change than multiple AND statements
	if !isValidAccountType(req.Type, vega.AccountType_ACCOUNT_TYPE_GENERAL, vega.AccountType_ACCOUNT_TYPE_MARGIN,
		vega.AccountType_ACCOUNT_TYPE_LOCK_WITHDRAW, vega.AccountType_ACCOUNT_TYPE_BOND, vega.AccountType_ACCOUNT_TYPE_UNSPECIFIED) {
		return nil, errors.New("invalid type for query, only GENERAL, MARGIN, LOCK_WITHDRAW AND BOND accounts for a party supported")
	}

	pagination := entities.OffsetPagination{}

	filter := entities.AccountFilter{
		Asset:        toAccountsFilterAsset(req.Asset),
		Parties:      toAccountsFilterParties(req.PartyId),
		AccountTypes: toAccountsFilterAccountTypes(req.Type),
		Markets:      toAccountsFilterMarkets(req.MarketId),
	}

	accountBalances, err := t.accountStore.QueryBalances(ctx, filter, pagination)
	if err != nil {
		return nil, apiError(codes.Internal, ErrAccountServiceGetPartyAccounts, err)
	}

	return &protoapi.PartyAccountsResponse{
		Accounts: accountBalancesToProtoAccountList(accountBalances),
	}, nil
}

func toAccountsFilterAccountTypes(accountTypes ...vega.AccountType) []vega.AccountType {
	accountTypesProto := make([]vega.AccountType, 0)

	for _, accountType := range accountTypes {
		if accountType == vega.AccountType_ACCOUNT_TYPE_UNSPECIFIED {
			return nil
		}

		accountTypesProto = append(accountTypesProto, accountType)
	}

	return accountTypesProto
}

func accountBalancesToProtoAccountList(accounts []entities.AccountBalance) []*vega.Account {
	accountsProto := make([]*vega.Account, 0, len(accounts))

	for _, acc := range accounts {
		accountsProto = append(accountsProto, acc.ToProto())
	}

	return accountsProto
}

func toAccountsFilterAsset(assetID string) entities.Asset {
	asset := entities.Asset{}

	if len(assetID) > 0 {
		asset.ID = entities.NewAssetID(assetID)
	}

	return asset
}

func toAccountsFilterParties(partyIDs ...string) []entities.Party {
	parties := make([]entities.Party, 0, len(partyIDs))
	for _, idStr := range partyIDs {
		if idStr == "" {
			continue
		}
		party := entities.Party{ID: entities.NewPartyID(idStr)}
		parties = append(parties, party)
	}

	return parties
}

func toAccountsFilterMarkets(marketIDs ...string) []entities.Market {
	markets := make([]entities.Market, 0, len(marketIDs))
	for _, idStr := range marketIDs {
		if idStr == "" {
			continue
		}
		market := entities.Market{ID: entities.NewMarketID(idStr)}
		markets = append(markets, market)
	}

	return markets
}

func (t *tradingDataDelegator) MarketAccounts(ctx context.Context,
	req *protoapi.MarketAccountsRequest,
) (*protoapi.MarketAccountsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("MarketAccounts")()

	filter := entities.AccountFilter{
		Asset:   toAccountsFilterAsset(req.Asset),
		Markets: toAccountsFilterMarkets(req.MarketId),
		AccountTypes: toAccountsFilterAccountTypes(
			vega.AccountType_ACCOUNT_TYPE_INSURANCE,
			vega.AccountType_ACCOUNT_TYPE_FEES_LIQUIDITY,
		),
	}

	pagination := entities.OffsetPagination{}

	accountBalances, err := t.accountStore.QueryBalances(ctx, filter, pagination)
	if err != nil {
		return nil, apiError(codes.Internal, ErrAccountServiceGetMarketAccounts, err)
	}

	return &protoapi.MarketAccountsResponse{
		Accounts: accountBalancesToProtoAccountList(accountBalances),
	}, nil
}

func (t *tradingDataDelegator) FeeInfrastructureAccounts(ctx context.Context,
	req *protoapi.FeeInfrastructureAccountsRequest,
) (*protoapi.FeeInfrastructureAccountsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("FeeInfrastructureAccounts")()

	filter := entities.AccountFilter{
		Asset: toAccountsFilterAsset(req.Asset),
		AccountTypes: toAccountsFilterAccountTypes(
			vega.AccountType_ACCOUNT_TYPE_FEES_INFRASTRUCTURE,
		),
	}
	pagination := entities.OffsetPagination{}

	accountBalances, err := t.accountStore.QueryBalances(ctx, filter, pagination)
	if err != nil {
		return nil, apiError(codes.Internal, ErrAccountServiceGetFeeInfrastructureAccounts, err)
	}
	return &protoapi.FeeInfrastructureAccountsResponse{
		Accounts: accountBalancesToProtoAccountList(accountBalances),
	}, nil
}

func (t *tradingDataDelegator) GlobalRewardPoolAccounts(ctx context.Context,
	req *protoapi.GlobalRewardPoolAccountsRequest,
) (*protoapi.GlobalRewardPoolAccountsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GloabRewardPoolAccounts")()
	filter := entities.AccountFilter{
		Asset: toAccountsFilterAsset(req.Asset),
		AccountTypes: toAccountsFilterAccountTypes(
			vega.AccountType_ACCOUNT_TYPE_GLOBAL_REWARD,
		),
	}
	pagination := entities.OffsetPagination{}

	accountBalances, err := t.accountStore.QueryBalances(ctx, filter, pagination)
	if err != nil {
		return nil, apiError(codes.Internal, ErrAccountServiceGetGlobalRewardPoolAccounts, err)
	}
	return &protoapi.GlobalRewardPoolAccountsResponse{
		Accounts: accountBalancesToProtoAccountList(accountBalances),
	}, nil
}

// MarketDataByID provides market data for the given ID.
func (t *tradingDataDelegator) MarketDataByID(ctx context.Context, req *protoapi.MarketDataByIDRequest) (*protoapi.MarketDataByIDResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("MarketDataByID_SQL")()

	// validate the market exist
	if req.MarketId != "" {
		_, err := t.marketsStore.GetByID(ctx, req.MarketId)
		if err != nil {
			return nil, apiError(codes.InvalidArgument, ErrInvalidMarketID, err)
		}
	}

	md, err := t.marketDataStore.GetMarketDataByID(ctx, req.MarketId)
	if err != nil {
		return nil, apiError(codes.Internal, ErrMarketServiceGetMarketData, err)
	}
	return &protoapi.MarketDataByIDResponse{
		MarketData: md.ToProto(),
	}, nil
}

// MarketsData provides all market data for all markets on this network.
func (t *tradingDataDelegator) MarketsData(ctx context.Context, _ *protoapi.MarketsDataRequest) (*protoapi.MarketsDataResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("MarketsData_SQL")()
	mds, _ := t.marketDataStore.GetMarketsData(ctx)

	mdptrs := make([]*vega.MarketData, 0, len(mds))
	for _, v := range mds {
		mdptrs = append(mdptrs, v.ToProto())
	}

	return &protoapi.MarketsDataResponse{
		MarketsData: mdptrs,
	}, nil
}

// MarketByID provides the given market.
func (t *tradingDataDelegator) MarketByID(ctx context.Context, req *protoapi.MarketByIDRequest) (*protoapi.MarketByIDResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("MarketByID_SQL")()

	mkt, err := validateMarketSQL(ctx, req.MarketId, t.marketsStore)
	if err != nil {
		return nil, err // validateMarket already returns an API error, no need to additionally wrap
	}

	return &protoapi.MarketByIDResponse{
		Market: mkt,
	}, nil
}

func validateMarketSQL(ctx context.Context, marketID string, marketsStore *sqlstore.Markets) (*vega.Market, error) {
	if len(marketID) == 0 {
		return nil, apiError(codes.InvalidArgument, ErrEmptyMissingMarketID)
	}

	market, err := marketsStore.GetByID(ctx, marketID)
	if err != nil {
		// We return nil for error as we do not want
		// to return an error when a market is not found
		// but just a nil value.
		return nil, nil
	}

	mkt, err := market.ToProto()
	if err != nil {
		return nil, nil
	}

	return mkt, nil
}

// Markets provides a list of all current markets that exist on the VEGA platform.
func (t *tradingDataDelegator) Markets(ctx context.Context, _ *protoapi.MarketsRequest) (*protoapi.MarketsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("Markets_SQL")()
	markets, err := t.marketsStore.GetAll(ctx, entities.OffsetPagination{})
	if err != nil {
		return nil, apiError(codes.Internal, ErrMarketServiceGetMarkets, err)
	}

	results := make([]*vega.Market, 0, len(markets))
	for _, m := range markets {
		mkt, err := m.ToProto()
		if err != nil {
			continue
		}

		results = append(results, mkt)
	}

	return &protoapi.MarketsResponse{
		Markets: results,
	}, nil
}

func (t *tradingDataDelegator) Deposit(ctx context.Context, req *protoapi.DepositRequest) (*protoapi.DepositResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("Deposit SQL")()
	if len(req.Id) <= 0 {
		return nil, ErrMissingDepositID
	}
	deposit, err := t.depositsStore.GetByID(ctx, req.Id)
	if err != nil {
		return nil, apiError(codes.NotFound, err)
	}
	return &protoapi.DepositResponse{
		Deposit: deposit.ToProto(),
	}, nil
}

func (t *tradingDataDelegator) Deposits(ctx context.Context, req *protoapi.DepositsRequest) (*protoapi.DepositsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("Deposits SQL")()
	if len(req.PartyId) <= 0 {
		return nil, ErrMissingPartyID
	}

	// current API doesn't support pagination, but we will need to support it for v2
	deposits := t.depositsStore.GetByParty(ctx, req.PartyId, false, entities.OffsetPagination{})
	out := make([]*vega.Deposit, 0, len(deposits))
	for _, v := range deposits {
		out = append(out, v.ToProto())
	}
	return &protoapi.DepositsResponse{
		Deposits: out,
	}, nil
}

func (t *tradingDataDelegator) EstimateMargin(ctx context.Context, req *protoapi.EstimateMarginRequest) (*protoapi.EstimateMarginResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("EstimateMargin SQL")()
	if req.Order == nil {
		return nil, apiError(codes.InvalidArgument, errors.New("missing order"))
	}

	margin, err := t.estimateMargin(ctx, req.Order)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	return &protoapi.EstimateMarginResponse{
		MarginLevels: margin,
	}, nil
}

func (t *tradingDataDelegator) estimateMargin(ctx context.Context, order *vega.Order) (*vega.MarginLevels, error) {
	if order.Side == vega.Side_SIDE_UNSPECIFIED {
		return nil, risk.ErrInvalidOrderSide
	}

	// first get the risk factors and market data (marketdata->markprice)
	rf, err := t.riskFactorStore.GetMarketRiskFactors(ctx, order.MarketId)
	if err != nil {
		return nil, err
	}
	mkt, err := t.marketsStore.GetByID(ctx, order.MarketId)
	if err != nil {
		return nil, err
	}

	mktProto, err := mkt.ToProto()
	if err != nil {
		return nil, err
	}

	mktData, err := t.marketDataStore.GetMarketDataByID(ctx, order.MarketId)
	if err != nil {
		return nil, err
	}

	f, err := num.DecimalFromString(rf.Short.String())
	if err != nil {
		return nil, err
	}
	if order.Side == vega.Side_SIDE_BUY {
		f, err = num.DecimalFromString(rf.Long.String())
		if err != nil {
			return nil, err
		}
	}

	asset, err := mktProto.GetAsset()
	if err != nil {
		return nil, err
	}

	// now calculate margin maintenance
	markPrice, _ := num.DecimalFromString(mktData.MarkPrice.String())

	// if the order is a limit order, use the limit price to calculate the margin maintenance
	if order.Type == vega.Order_TYPE_LIMIT {
		markPrice, _ = num.DecimalFromString(order.Price)
	}

	maintenanceMargin := num.DecimalFromFloat(float64(order.Size)).Mul(f).Mul(markPrice)
	// now we use the risk factors
	return &vega.MarginLevels{
		PartyId:                order.PartyId,
		MarketId:               mktProto.GetId(),
		Asset:                  asset,
		Timestamp:              0,
		MaintenanceMargin:      maintenanceMargin.String(),
		SearchLevel:            maintenanceMargin.Mul(num.DecimalFromFloat(mkt.TradableInstrument.MarginCalculator.ScalingFactors.SearchLevel)).String(),
		InitialMargin:          maintenanceMargin.Mul(num.DecimalFromFloat(mkt.TradableInstrument.MarginCalculator.ScalingFactors.InitialMargin)).String(),
		CollateralReleaseLevel: maintenanceMargin.Mul(num.DecimalFromFloat(mkt.TradableInstrument.MarginCalculator.ScalingFactors.CollateralRelease)).String(),
	}, nil
}

func (t *tradingDataDelegator) EstimateFee(ctx context.Context, req *protoapi.EstimateFeeRequest) (*protoapi.EstimateFeeResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("EstimateFee SQL")()
	if req.Order == nil {
		return nil, apiError(codes.InvalidArgument, errors.New("missing order"))
	}

	fee, err := t.estimateFee(ctx, req.Order)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	return &protoapi.EstimateFeeResponse{
		Fee: fee,
	}, nil
}

func (t *tradingDataDelegator) estimateFee(ctx context.Context, order *vega.Order) (*vega.Fee, error) {
	mkt, err := t.marketsStore.GetByID(ctx, order.MarketId)
	if err != nil {
		return nil, err
	}
	price, overflowed := num.UintFromString(order.Price, 10)
	if overflowed {
		return nil, errors.New("invalid order price")
	}
	if order.PeggedOrder != nil {
		return &vega.Fee{
			MakerFee:          "0",
			InfrastructureFee: "0",
			LiquidityFee:      "0",
		}, nil
	}

	base := num.DecimalFromUint(price.Mul(price, num.NewUint(order.Size)))
	maker, infra, liquidity, err := t.feeFactors(mkt)
	if err != nil {
		return nil, err
	}

	fee := &vega.Fee{
		MakerFee:          base.Mul(num.NewDecimalFromFloat(maker)).String(),
		InfrastructureFee: base.Mul(num.NewDecimalFromFloat(infra)).String(),
		LiquidityFee:      base.Mul(num.NewDecimalFromFloat(liquidity)).String(),
	}

	return fee, nil
}

func (t *tradingDataDelegator) feeFactors(mkt entities.Market) (maker, infra, liquidity float64, err error) {
	if maker, err = strconv.ParseFloat(mkt.Fees.Factors.MakerFee, 64); err != nil {
		return
	}
	if infra, err = strconv.ParseFloat(mkt.Fees.Factors.InfrastructureFee, 64); err != nil {
		return
	}
	if liquidity, err = strconv.ParseFloat(mkt.Fees.Factors.LiquidityFee, 64); err != nil {
		return
	}

	return
}

// MarginLevels returns the current margin levels for a given party and market.
func (t *tradingDataDelegator) MarginLevels(ctx context.Context, req *protoapi.MarginLevelsRequest) (*protoapi.MarginLevelsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("MarginLevels SQL")()

	mls, err := t.marginLevelsStore.GetMarginLevelsByID(ctx, req.PartyId, req.MarketId, entities.OffsetPagination{})
	if err != nil {
		return nil, apiError(codes.Internal, ErrRiskServiceGetMarginLevelsByID, err)
	}
	levels := make([]*vega.MarginLevels, 0, len(mls))
	for _, v := range mls {
		proto, err := v.ToProto(t.accountStore)
		if err != nil {
			return nil, apiError(codes.Internal, ErrRiskServiceGetMarginLevelsByID, err)
		}
		levels = append(levels, proto)
	}
	return &protoapi.MarginLevelsResponse{
		MarginLevels: levels,
	}, nil
}

func (t *tradingDataDelegator) GetRiskFactors(ctx context.Context, in *protoapi.GetRiskFactorsRequest) (*protoapi.GetRiskFactorsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetRiskFactors SQL")()

	rfs, err := t.riskFactorStore.GetMarketRiskFactors(ctx, in.MarketId)
	if err != nil {
		return nil, nil
	}

	return &protoapi.GetRiskFactorsResponse{
		RiskFactor: rfs.ToProto(),
	}, nil
}

func (t *tradingDataDelegator) Withdrawal(ctx context.Context, req *protoapi.WithdrawalRequest) (*protoapi.WithdrawalResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("Withdrawal SQL")()
	if len(req.Id) <= 0 {
		return nil, ErrMissingDepositID
	}
	withdrawal, err := t.withdrawalsStore.GetByID(ctx, req.Id)
	if err != nil {
		return nil, apiError(codes.NotFound, err)
	}
	return &protoapi.WithdrawalResponse{
		Withdrawal: withdrawal.ToProto(),
	}, nil
}

func (t *tradingDataDelegator) Withdrawals(ctx context.Context, req *protoapi.WithdrawalsRequest) (*protoapi.WithdrawalsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("Withdrawals SQL")()
	if len(req.PartyId) <= 0 {
		return nil, ErrMissingPartyID
	}

	// current API doesn't support pagination, but we will need to support it for v2
	withdrawals := t.withdrawalsStore.GetByParty(ctx, req.PartyId, false, entities.OffsetPagination{})
	out := make([]*vega.Withdrawal, 0, len(withdrawals))
	for _, w := range withdrawals {
		out = append(out, w.ToProto())
	}
	return &protoapi.WithdrawalsResponse{
		Withdrawals: out,
	}, nil
}

func (t *tradingDataDelegator) ERC20WithdrawalApproval(ctx context.Context, req *protoapi.ERC20WithdrawalApprovalRequest) (*protoapi.ERC20WithdrawalApprovalResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("ERC20WithdrawalApproval SQL")()
	if len(req.WithdrawalId) <= 0 {
		return nil, ErrMissingDepositID
	}

	// get withdrawal first
	w, err := t.withdrawalsStore.GetByID(ctx, req.WithdrawalId)
	if err != nil {
		return nil, apiError(codes.NotFound, err)
	}

	// get the signatures from  notaryStore
	signatures, err := t.notaryStore.GetByResourceID(ctx, req.WithdrawalId)
	if err != nil {
		return nil, apiError(codes.NotFound, err)
	}

	// some assets stuff
	assets, err := t.assetStore.GetAll(ctx)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	// get the signature into the form form:
	// 0x + sig1 + sig2 + ... + sigN in hex encoded form
	pack := "0x"
	for _, v := range signatures {
		pack = fmt.Sprintf("%v%v", pack, hex.EncodeToString(v.Sig))
	}

	var address string
	for _, v := range assets {
		if v.ID == w.Asset {
			address = v.ERC20Contract
			break // found the one we want
		}
	}
	if len(address) <= 0 {
		return nil, fmt.Errorf("invalid erc20 token contract address")
	}

	return &protoapi.ERC20WithdrawalApprovalResponse{
		AssetSource:   address,
		Amount:        fmt.Sprintf("%v", w.Amount),
		Expiry:        w.Expiry.UnixMicro(),
		Nonce:         w.Ref,
		TargetAddress: w.Ext.GetErc20().ReceiverAddress,
		Signatures:    pack,
		// timestamps is unix nano, contract needs unix. So load if first, and cut nanos
		Creation: w.CreatedTimestamp.Unix(),
	}, nil
}

func (t *tradingDataDelegator) GetNodeSignaturesAggregate(ctx context.Context,
	req *protoapi.GetNodeSignaturesAggregateRequest,
) (*protoapi.GetNodeSignaturesAggregateResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetNodeSignaturesAggregate SQL")()
	if len(req.Id) <= 0 {
		return nil, apiError(codes.InvalidArgument, errors.New("missing ID"))
	}

	sigs, err := t.notaryStore.GetByResourceID(ctx, req.Id)
	if err != nil {
		return nil, apiError(codes.NotFound, err)
	}

	out := make([]*commandspb.NodeSignature, 0, len(sigs))
	for _, v := range sigs {
		vv := v.ToProto()
		out = append(out, vv)
	}

	return &protoapi.GetNodeSignaturesAggregateResponse{
		Signatures: out,
	}, nil
}

func (t *tradingDataDelegator) OracleSpec(ctx context.Context, req *protoapi.OracleSpecRequest) (*protoapi.OracleSpecResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("OracleSpec SQL")()
	if len(req.Id) <= 0 {
		return nil, ErrMissingOracleSpecID
	}
	spec, err := t.oracleSpecStore.GetSpecByID(ctx, req.Id)
	if err != nil {
		return nil, apiError(codes.NotFound, err)
	}
	return &protoapi.OracleSpecResponse{
		OracleSpec: spec.ToProto(),
	}, nil
}

func (t *tradingDataDelegator) OracleSpecs(ctx context.Context, _ *protoapi.OracleSpecsRequest) (*protoapi.OracleSpecsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("OracleSpecs SQL")()
	specs, err := t.oracleSpecStore.GetSpecs(ctx, entities.OffsetPagination{})
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	out := make([]*oraclespb.OracleSpec, 0, len(specs))
	for _, v := range specs {
		out = append(out, v.ToProto())
	}

	return &protoapi.OracleSpecsResponse{
		OracleSpecs: out,
	}, nil
}

func (t *tradingDataDelegator) OracleDataBySpec(ctx context.Context, req *protoapi.OracleDataBySpecRequest) (*protoapi.OracleDataBySpecResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("OracleDataBySpec SQL")()
	if len(req.Id) <= 0 {
		return nil, ErrMissingOracleSpecID
	}
	data, err := t.oracleDataStore.GetOracleDataBySpecID(ctx, req.Id, entities.OffsetPagination{})
	if err != nil {
		return nil, apiError(codes.NotFound, err)
	}
	out := make([]*oraclespb.OracleData, 0, len(data))
	for _, v := range data {
		out = append(out, v.ToProto())
	}
	return &protoapi.OracleDataBySpecResponse{
		OracleData: out,
	}, nil
}

func (t *tradingDataDelegator) ListOracleData(ctx context.Context, _ *protoapi.ListOracleDataRequest) (*protoapi.ListOracleDataResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("ListOracleData SQL")()
	specs, err := t.oracleDataStore.ListOracleData(ctx, entities.OffsetPagination{})
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	out := make([]*oraclespb.OracleData, 0, len(specs))
	for _, v := range specs {
		out = append(out, v.ToProto())
	}

	return &protoapi.ListOracleDataResponse{
		OracleData: out,
	}, nil
}

func (t *tradingDataDelegator) LiquidityProvisions(ctx context.Context, req *protoapi.LiquidityProvisionsRequest) (*protoapi.LiquidityProvisionsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("LiquidityProvisions")()

	partyID := entities.NewPartyID(req.Party)
	marketID := entities.NewMarketID(req.Market)

	lps, err := t.liquidityProvisionStore.Get(ctx, partyID, marketID, entities.OffsetPagination{})
	if err != nil {
		return nil, err
	}

	out := make([]*vega.LiquidityProvision, 0, len(lps))
	for _, v := range lps {
		out = append(out, v.ToProto())
	}
	return &protoapi.LiquidityProvisionsResponse{
		LiquidityProvisions: out,
	}, nil
}

func (t *tradingDataDelegator) PartyStake(ctx context.Context, req *protoapi.PartyStakeRequest) (*protoapi.PartyStakeResponse, error) {
	if len(req.Party) <= 0 {
		return nil, apiError(codes.InvalidArgument, errors.New("missing party id"))
	}

	partyID := entities.NewPartyID(req.Party)

	stake, stakeLinkings := t.stakingStore.GetStake(ctx, partyID, entities.OffsetPagination{})
	outStakeLinkings := make([]*eventspb.StakeLinking, 0, len(stakeLinkings))
	for _, v := range stakeLinkings {
		outStakeLinkings = append(outStakeLinkings, v.ToProto())
	}

	return &protoapi.PartyStakeResponse{
		CurrentStakeAvailable: num.UintToString(stake),
		StakeLinkings:         outStakeLinkings,
	}, nil
}

func (t *tradingDataDelegator) GetKeyRotations(ctx context.Context, req *protoapi.GetKeyRotationsRequest) (*protoapi.GetKeyRotationsResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetKeyRotations")()

	rotations, err := t.keyRotationsStore.GetAllPubKeyRotations(ctx)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	protoRotations := make([]*protoapi.KeyRotation, len(rotations))
	for i, v := range rotations {
		protoRotations[i] = v.ToProto()
	}

	return &protoapi.GetKeyRotationsResponse{
		Rotations: protoRotations,
	}, nil
}

func (t *tradingDataDelegator) GetKeyRotationsByNode(ctx context.Context, req *protoapi.GetKeyRotationsByNodeRequest) (*protoapi.GetKeyRotationsByNodeResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetKeyRotationsByNode")()

	if req.GetNodeId() == "" {
		return nil, apiError(codes.InvalidArgument, errors.New("missing node ID parameter"))
	}

	rotations, err := t.keyRotationsStore.GetPubKeyRotationsPerNode(ctx, req.GetNodeId())
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	protoRotations := make([]*protoapi.KeyRotation, len(rotations))
	for i, v := range rotations {
		protoRotations[i] = v.ToProto()
	}

	return &protoapi.GetKeyRotationsByNodeResponse{
		Rotations: protoRotations,
	}, nil
}

func (t *tradingDataDelegator) GetNodeData(ctx context.Context, req *protoapi.GetNodeDataRequest) (*protoapi.GetNodeDataResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetNodeData")()

	nodeData, err := t.nodeStore.GetNodeData(ctx)
	if err != nil {
		return nil, apiError(codes.Internal, err)
	}

	return &protoapi.GetNodeDataResponse{
		NodeData: nodeData.ToProto(),
	}, nil
}

func (t *tradingDataDelegator) GetNodes(ctx context.Context, req *protoapi.GetNodesRequest) (*protoapi.GetNodesResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetNodes")()

	epoch, err := t.epochStore.GetCurrent(ctx)
	if err != nil {
		fmt.Printf("%v", err)
		return nil, apiError(codes.Internal, err)
	}

	nodes, err := t.nodeStore.GetNodes(ctx, uint64(epoch.ID))
	if err != nil {
		fmt.Printf("%v", err)
		return nil, apiError(codes.Internal, err)
	}

	protoNodes := make([]*vega.Node, len(nodes))
	for i, v := range nodes {
		protoNodes[i] = v.ToProto()
	}

	return &protoapi.GetNodesResponse{
		Nodes: protoNodes,
	}, nil
}

func (t *tradingDataDelegator) GetNodeByID(ctx context.Context, req *protoapi.GetNodeByIDRequest) (*protoapi.GetNodeByIDResponse, error) {
	defer metrics.StartAPIRequestAndTimeGRPC("GetNodeByID")()

	if req.GetId() == "" {
		return nil, apiError(codes.InvalidArgument, errors.New("missing node ID parameter"))
	}

	epoch, err := t.epochStore.GetCurrent(ctx)
	if err != nil {
		fmt.Printf("%v", err)
		return nil, apiError(codes.Internal, err)
	}

	node, err := t.nodeStore.GetNodeByID(ctx, req.GetId(), uint64(epoch.ID))
	if err != nil {
		fmt.Printf("%v", err)
		return nil, apiError(codes.NotFound, err)
	}

	return &protoapi.GetNodeByIDResponse{
		Node: node.ToProto(),
	}, nil
}
