package entities

import (
	"fmt"
	"strconv"
	"time"

	"code.vegaprotocol.io/protos/vega"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"github.com/shopspring/decimal"
)

type Reward struct {
	PartyID        []byte
	AssetID        []byte
	EpochID        int64
	Amount         decimal.Decimal
	PercentOfTotal float64
	VegaTime       time.Time
}

func (r *Reward) PartyHexID() string {
	return Party{ID: r.PartyID}.HexID()
}

func (r *Reward) AssetHexID() string {
	return Asset{ID: r.AssetID}.HexID()
}

func (r Reward) String() string {
	return fmt.Sprintf("{Epoch: %v, Party: %s, Asset: %s, Amount: %v}",
		r.EpochID, r.PartyHexID(), r.AssetHexID(), r.Amount)
}

func (r *Reward) ToProto() *vega.Reward {
	protoReward := vega.Reward{
		PartyId:           r.PartyHexID(),
		AssetId:           r.AssetHexID(),
		Epoch:             uint64(r.EpochID),
		Amount:            r.Amount.String(),
		PercentageOfTotal: fmt.Sprintf("%v", r.PercentOfTotal),
		ReceivedAt:        r.VegaTime.UnixNano(),
	}
	return &protoReward
}

func RewardFromProto(pr eventspb.RewardPayoutEvent) (Reward, error) {
	partyID, err := MakePartyID(pr.Party)
	if err != nil {
		return Reward{}, fmt.Errorf("parsing party id '%v': %w", pr.Party, err)
	}

	assetID := MakeAssetID(pr.Asset)

	epochID, err := strconv.ParseInt(pr.EpochSeq, 10, 64)
	if err != nil {
		return Reward{}, fmt.Errorf("parsing epoch '%v': %w", pr.EpochSeq, err)
	}

	percentOfTotal, err := strconv.ParseFloat(pr.PercentOfTotalReward, 64)
	if err != nil {
		return Reward{}, fmt.Errorf("parsing percent of total reward '%v': %w",
			pr.PercentOfTotalReward, err)
	}

	reward := Reward{
		PartyID:        partyID,
		AssetID:        assetID,
		EpochID:        epochID,
		PercentOfTotal: percentOfTotal,
		VegaTime:       time.Unix(0, pr.Timestamp),
	}

	return reward, nil
}
