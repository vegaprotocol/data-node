package gql

import (
	"context"

	v1 "code.vegaprotocol.io/protos/data-node/api/v1"
	"code.vegaprotocol.io/protos/vega"
)

type rewardPerAssetDetailResolver VegaResolverRoot

func (r *rewardPerAssetDetailResolver) Asset(ctx context.Context, obj *vega.RewardPerAssetDetail) (*vega.Asset, error) {
	asset, err := r.r.getAssetByID(ctx, obj.Asset)
	if err != nil {
		return nil, err
	}

	return asset, nil
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func paginateRewards(rewards []*vega.RewardDetails, p *v1.Pagination) []*vega.RewardDetails {
	length := uint64(len(rewards))
	start := uint64(0)
	end := length

	if p.Descending {
		if p.Skip+p.Limit <= length {
			start = length - p.Skip - p.Limit
		}
		end = length - min(p.Skip, length)
	} else {
		start = p.Skip
		if p.Limit != 0 {
			end = p.Skip + p.Limit
		}
	}

	start = min(start, length)
	end = min(end, length)
	return rewards[start:end]
}

func (r *rewardPerAssetDetailResolver) Rewards(ctx context.Context, obj *vega.RewardPerAssetDetail, skip, first, last *int) ([]*vega.RewardDetails, error) {
	p := makePagination(skip, first, last)
	return paginateRewards(obj.Details, p), nil
}

func (r *rewardPerAssetDetailResolver) TotalAmount(ctx context.Context, obj *vega.RewardPerAssetDetail) (string, error) {
	return obj.TotalForAsset, nil
}
