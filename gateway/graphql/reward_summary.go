package gql

import (
	"context"

	protoapi "code.vegaprotocol.io/protos/data-node/api/v1"
	"code.vegaprotocol.io/protos/vega"
)

type rewardSummaryResolver VegaResolverRoot

func (r *rewardSummaryResolver) Asset(ctx context.Context, obj *vega.RewardSummary) (*vega.Asset, error) {
	return r.r.getAssetByID(ctx, obj.AssetId)
}

func (r *rewardSummaryResolver) Rewards(ctx context.Context, obj *vega.RewardSummary, skip, first, last *int) ([]*vega.Reward, error) {
	p := makePagination(skip, first, last)

	req := &protoapi.GetRewardsRequest{
		PartyId:    obj.PartyId,
		AssetId:    obj.AssetId,
		Pagination: p,
	}
	resp, err := r.tradingDataClient.GetRewards(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Rewards, nil
}
