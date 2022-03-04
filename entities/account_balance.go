package entities

import (
	"encoding/hex"
	"strings"

	"code.vegaprotocol.io/protos/vega"
	"github.com/shopspring/decimal"
)

type AccountBalance struct {
	*Account
	Balance decimal.Decimal
}

func (ab *AccountBalance) ToProto() *vega.Account {
	owner := ""
	market := ""

	if len(ab.PartyID) > 0 {
		owner = hex.EncodeToString(ab.PartyID)
	}

	var assetID string
	if assetID = string(ab.AssetID); strings.HasPrefix(assetID, badAssetPrefix) {
		assetID = strings.TrimPrefix(assetID, badAssetPrefix)
	} else {
		assetID = hex.EncodeToString(ab.AssetID)
	}

	if len(ab.MarketID) > 0 {
		market = hex.EncodeToString(ab.MarketID)
	}

	return &vega.Account{
		Owner:    owner,
		Balance:  ab.Balance.String(),
		Asset:    assetID,
		MarketId: market,
		Type:     ab.Account.Type,
	}
}
