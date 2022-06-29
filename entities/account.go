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

package entities

import (
	"fmt"
	"time"

	"code.vegaprotocol.io/protos/vega"
)

const (
	noMarketStr    string = "!"
	systemOwnerStr string = "*"
)

type Account struct {
	ID       int64
	PartyID  PartyID
	AssetID  AssetID
	MarketID MarketID
	Type     vega.AccountType
	VegaTime time.Time
}

func (a Account) String() string {
	return fmt.Sprintf("{ID: %s}", a.AssetID)
}

func AccountFromProto(va *vega.Account) (Account, error) {
	account := Account{
		PartyID:  NewPartyID(va.Owner),
		AssetID:  NewAssetID(va.Asset),
		MarketID: NewMarketID(va.MarketId),
		Type:     va.Type,
	}
	return account, nil
}

func AccountFromAccountID(id *vega.AccountId) Account {
	return Account{
		PartyID:  NewPartyID(id.Owner),
		AssetID:  NewAssetID(id.Asset),
		MarketID: NewMarketID(id.MarketId),
		Type:     id.Type,
	}
}
