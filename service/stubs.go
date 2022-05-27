package service

import "code.vegaprotocol.io/data-node/sqlstore"

type Asset struct{ *sqlstore.Assets }
type Block struct{ *sqlstore.Blocks }
type Party struct{ *sqlstore.Parties }
type Order struct{ *sqlstore.Orders }
type NetworkLimits struct{ *sqlstore.NetworkLimits }
type Markets struct{ *sqlstore.Markets }
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
