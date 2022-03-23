package entities

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"code.vegaprotocol.io/protos/vega"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/encoding/protojson"
)

type Withdrawal struct {
	ID                 []byte
	PartyID            []byte
	Amount             decimal.Decimal
	Asset              []byte
	Status             WithdrawalStatus
	Ref                string
	Expiry             time.Time
	TxHash             string
	CreatedTimestamp   time.Time
	WithdrawnTimestamp time.Time
	Ext                WithdrawExt
	VegaTime           time.Time
}

// XXX: Remove when IDs have their own type
func MakeWithdrawalID(stringID string) ([]byte, error) {
	id, err := hex.DecodeString(stringID)
	if err != nil {
		return nil, fmt.Errorf("id is not a valid hex string: %s", stringID)
	}
	return id, nil
}

func WithdrawalFromProto(withdrawal *vega.Withdrawal, vegaTime time.Time) (*Withdrawal, error) {
	var id, partyID []byte
	var err error
	var amount decimal.Decimal

	if id, err = decodeID(withdrawal.Id); err != nil {
		return nil, fmt.Errorf("invalid withdrawal id: %v", err)
	}
	if partyID, err = decodeID(withdrawal.PartyId); err != nil {
		return nil, fmt.Errorf("invalid party id: %w", err)
	}

	if amount, err = decimal.NewFromString(withdrawal.Amount); err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	return &Withdrawal{
		ID:                 id,
		PartyID:            partyID,
		Amount:             amount,
		Asset:              MakeAssetID(withdrawal.Asset),
		Status:             WithdrawalStatus(withdrawal.Status),
		Ref:                withdrawal.Ref,
		Expiry:             time.Unix(0, withdrawal.Expiry),
		TxHash:             withdrawal.TxHash,
		CreatedTimestamp:   time.Unix(0, withdrawal.CreatedTimestamp),
		WithdrawnTimestamp: time.Unix(0, withdrawal.WithdrawnTimestamp),
		Ext:                WithdrawExt{withdrawal.Ext},
		VegaTime:           vegaTime,
	}, nil
}

func (w Withdrawal) HexID() string {
	return hex.EncodeToString(w.ID)
}

func (w Withdrawal) ToProto() *vega.Withdrawal {
	assetID := hex.EncodeToString(w.Asset)

	if strings.HasPrefix(string(w.Asset), badAssetPrefix) {
		assetID = strings.TrimPrefix(string(w.Asset), badAssetPrefix)
	}

	return &vega.Withdrawal{
		Id:                 hex.EncodeToString(w.ID),
		PartyId:            hex.EncodeToString(w.PartyID),
		Amount:             w.Amount.String(),
		Asset:              assetID,
		Status:             vega.Withdrawal_Status(w.Status),
		Ref:                w.Ref,
		Expiry:             w.Expiry.UnixNano(),
		TxHash:             w.TxHash,
		CreatedTimestamp:   w.CreatedTimestamp.UnixNano(),
		WithdrawnTimestamp: w.WithdrawnTimestamp.UnixNano(),
		Ext:                w.Ext.WithdrawExt,
	}
}

type WithdrawExt struct {
	*vega.WithdrawExt
}

func (we WithdrawExt) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(we)
}

func (we *WithdrawExt) UnmarshalJSON(b []byte) error {
	we.WithdrawExt = &vega.WithdrawExt{}
	return protojson.Unmarshal(b, we)
}
