package entities

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"code.vegaprotocol.io/protos/vega"
	"github.com/shopspring/decimal"
)

type Deposit struct {
	ID                []byte
	Status            DepositStatus
	PartyID           []byte
	Asset             []byte
	Amount            decimal.Decimal
	TxHash            []byte
	CreditedTimestamp time.Time
	CreatedTimestamp  time.Time
	VegaTime          time.Time
}

func makeID(stringID string) ([]byte, error) {
	// TODO: Check why we are receiving a hash with the 0x prefix, could be a bug in core
	// 		 but for now just remove it so we can write the data in the database as a byte array
	stringID = strings.TrimPrefix(stringID, "0x")

	id, err := hex.DecodeString(stringID)
	if err != nil {
		return nil, fmt.Errorf("id is not a valid hex string: %s", stringID)
	}
	return id, nil
}

func DepositFromProto(deposit *vega.Deposit, vegaTime time.Time) (*Deposit, error) {
	var id, partyID, txHash []byte
	var err error
	var amount decimal.Decimal

	if id, err = makeID(deposit.Id); err != nil {
		return nil, fmt.Errorf("invalid deposit id: %v", err)
	}
	if partyID, err = makeID(deposit.PartyId); err != nil {
		return nil, fmt.Errorf("invalid party id: %w", err)
	}
	if txHash, err = makeID(deposit.TxHash); err != nil {
		return nil, fmt.Errorf("invalid transaction hash: %w", err)
	}
	if amount, err = decimal.NewFromString(deposit.Amount); err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	return &Deposit{
		ID:                id,
		Status:            DepositStatus(deposit.Status),
		PartyID:           partyID,
		Asset:             MakeAssetID(deposit.Asset),
		Amount:            amount,
		TxHash:            txHash,
		CreditedTimestamp: time.Unix(0, deposit.CreditedTimestamp),
		CreatedTimestamp:  time.Unix(0, deposit.CreatedTimestamp),
		VegaTime:          vegaTime,
	}, nil
}

func (d Deposit) HexID() string {
	return hex.EncodeToString(d.ID)
}

func (d Deposit) ToProto() *vega.Deposit {
	assetID := hex.EncodeToString(d.Asset)

	if strings.HasPrefix(string(d.Asset), badAssetPrefix) {
		assetID = strings.TrimPrefix(string(d.Asset), badAssetPrefix)
	}
	return &vega.Deposit{
		Id:                hex.EncodeToString(d.ID),
		Status:            vega.Deposit_Status(d.Status),
		PartyId:           hex.EncodeToString(d.PartyID),
		Asset:             assetID,
		Amount:            d.Amount.String(),
		TxHash:            hex.EncodeToString(d.TxHash),
		CreditedTimestamp: d.CreditedTimestamp.UnixNano(),
		CreatedTimestamp:  d.CreatedTimestamp.UnixNano(),
	}
}
