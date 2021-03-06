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

package service

import (
	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/data-node/sqlstore"
)

type Asset struct{ *sqlstore.Assets }
type Block struct{ *sqlstore.Blocks }
type Party struct{ *sqlstore.Parties }
type NetworkLimits struct{ *sqlstore.NetworkLimits }
type Epoch struct{ *sqlstore.Epochs }
type Deposit struct{ *sqlstore.Deposits }
type Withdrawal struct{ *sqlstore.Withdrawals }
type RiskFactor struct{ *sqlstore.RiskFactors }
type NetworkParameter struct{ *sqlstore.NetworkParameters }
type Checkpoint struct{ *sqlstore.Checkpoints }
type OracleSpec struct{ *sqlstore.OracleSpec }
type OracleData struct{ *sqlstore.OracleData }
type LiquidityProvision struct{ *sqlstore.LiquidityProvision }
type Transfer struct{ *sqlstore.Transfers }
type StakeLinking struct{ *sqlstore.StakeLinking }
type Notary struct{ *sqlstore.Notary }
type MultiSig struct {
	*sqlstore.ERC20MultiSigSignerEvent
}
type KeyRotations struct{ *sqlstore.KeyRotations }
type Node struct{ *sqlstore.Node }

func NewAsset(store *sqlstore.Assets, log *logging.Logger) *Asset {
	return &Asset{Assets: store}
}

func NewBlock(store *sqlstore.Blocks, log *logging.Logger) *Block {
	return &Block{Blocks: store}
}

func NewParty(store *sqlstore.Parties, log *logging.Logger) *Party {
	return &Party{Parties: store}
}

func NewNetworkLimits(store *sqlstore.NetworkLimits, log *logging.Logger) *NetworkLimits {
	return &NetworkLimits{NetworkLimits: store}
}

func NewEpoch(store *sqlstore.Epochs, log *logging.Logger) *Epoch {
	return &Epoch{Epochs: store}
}

func NewDeposit(store *sqlstore.Deposits, log *logging.Logger) *Deposit {
	return &Deposit{Deposits: store}
}

func NewWithdrawal(store *sqlstore.Withdrawals, log *logging.Logger) *Withdrawal {
	return &Withdrawal{Withdrawals: store}
}

func NewRiskFactor(store *sqlstore.RiskFactors, log *logging.Logger) *RiskFactor {
	return &RiskFactor{RiskFactors: store}
}

func NewNetworkParameter(store *sqlstore.NetworkParameters, log *logging.Logger) *NetworkParameter {
	return &NetworkParameter{NetworkParameters: store}
}

func NewCheckpoint(store *sqlstore.Checkpoints, log *logging.Logger) *Checkpoint {
	return &Checkpoint{Checkpoints: store}
}

func NewOracleSpec(store *sqlstore.OracleSpec, log *logging.Logger) *OracleSpec {
	return &OracleSpec{OracleSpec: store}
}

func NewOracleData(store *sqlstore.OracleData, log *logging.Logger) *OracleData {
	return &OracleData{OracleData: store}
}

func NewLiquidityProvision(store *sqlstore.LiquidityProvision, log *logging.Logger) *LiquidityProvision {
	return &LiquidityProvision{LiquidityProvision: store}
}

func NewTransfer(store *sqlstore.Transfers, log *logging.Logger) *Transfer {
	return &Transfer{Transfers: store}
}

func NewStakeLinking(store *sqlstore.StakeLinking, log *logging.Logger) *StakeLinking {
	return &StakeLinking{StakeLinking: store}
}

func NewNotary(store *sqlstore.Notary, log *logging.Logger) *Notary {
	return &Notary{Notary: store}
}

func NewMultiSig(store *sqlstore.ERC20MultiSigSignerEvent, log *logging.Logger) *MultiSig {
	return &MultiSig{ERC20MultiSigSignerEvent: store}
}

func NewKeyRotations(store *sqlstore.KeyRotations, log *logging.Logger) *KeyRotations {
	return &KeyRotations{KeyRotations: store}
}

func NewNode(store *sqlstore.Node, log *logging.Logger) *Node {
	return &Node{Node: store}
}
