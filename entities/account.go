package entities

import (
	"time"

	"code.vegaprotocol.io/protos/vega"
)

type Account struct {
	ID       int64
	PartyID  []byte
	AssetID  []byte
	MarketID []byte
	Type     vega.AccountType
	VegaTime time.Time
}
