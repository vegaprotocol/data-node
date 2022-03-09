package integration_test

import (
	"strings"
	"time"
)

type Market struct {
	Id                            HexString
	Name                          string
	Fees                          Fees
	Trades                        []Trade
	TradableInstrument            TradableInstrument
	DecimalPlaces                 int
	OpeningAuction                ActionDuration
	PriceMonitoringSettings       PriceMonitoringSettings
	LiquidityMonitoringParameters LiquidityMonitoringParameters
	TradingMode                   string
	State                         string
	Proposal                      Proposal
	Orders                        []Order
	Accounts                      []Account
}

type Fees struct {
	MakerFee          string
	InfrastructureFee string
	LiquidityFee      string
}

type TradableInstrument struct {
	Instrument Instrument
	// RiskModel        RiskModel
	MarginCalculator MarginCalculator
}

type Instrument struct {
	Id       string
	Code     string
	Name     string
	Metadata Metadata
}

type MarginCalculator struct {
	ScalingFactors ScalingFactors
}

type ScalingFactors struct {
	SearchLevel       float64
	InitialMargin     float64
	CollateralRelease float64
}

type Metadata struct {
	Tags []string
}

type ActionDuration struct {
	DurationSecs int
	Volume       int
}
type Trade struct {
	Id                 HexString
	Price              string
	Size               string
	CreatedAt          TimeString
	Market             Market
	BuyOrder           HexString
	SellOrder          HexString
	Buyer              Party
	Seller             Party
	Aggressor          string
	Type               string
	BuyerFee           TradeFee
	SellerFee          TradeFee
	BuyerAuctionBatch  int
	SellerAuctionBatch int
}

type TradeFee struct {
	MakerFee          string
	InfrastructureFee string
	LiquidityFee      string
}
type PriceMonitoringSettings struct {
	Parameters          PriceMonitoringParameters
	UpdateFrequencySecs int
}

type PriceMonitoringParameters struct {
	Triggers []PriceMonitoringTrigger
}

type PriceMonitoringTrigger struct {
	HorizonSecs          int
	Probability          float64
	AuctionExtensionSecs float64
}

type LiquidityMonitoringParameters struct {
	TargetStakeParameters TargetStakeParameters
	TriggeringRatio       float64
}

type TargetStakeParameters struct {
	TimeWindow    int
	ScalingFactor float64
}

type Proposal struct {
	Id              HexString
	Reference       string
	Party           Party
	State           string
	Datetime        TimeString
	Terms           ProposalTerms
	Votes           ProposalVotes
	RejectionReason string
}

type Party struct {
	Id                 HexString
	Orders             []Order
	Trades             []Trade
	Accounts           []Account
	Proposals          []Proposal
	Votes              []Vote
	LiquidityProvision []LiquidityProvision
	// TODO:
	// Positions []Position
	// Margins []MarginLevels
	// Withdrawals []WithDrawal
	// Deposits []Deposit
	// Delegations []Delegation
	// Stake PartyStake
	// Rewards []Reward
	// RewardSummaries []RewardSummary

}

type ProposalTerms struct {
	ClosingDateTime   TimeString
	EnactmentDatetime TimeString
	// TODO: Change (can't to ...on Foo yet)
}

type ProposalVotes struct {
	Yes ProposalVoteSide
	No  ProposalVoteSide
}

type ProposalVoteSide struct {
	Votes       []Vote
	TotalNumber string
	TotalWeight string
	TotalTokens string
}

type Vote struct {
	Value                  string
	Party                  Party
	Datetime               TimeString
	ProposalId             HexString
	GovernanceTokenBalance string
	GovernanceTokenWeight  string
}

type Order struct {
	Id                 string
	Price              string
	Side               string
	Timeinforce        string
	Market             Market
	Size               string
	Remaining          string
	Party              Party
	CreatedAt          TimeString
	ExpiresAt          TimeString
	Status             string
	Reference          string
	Trades             []Trade
	Type               string
	RejectionReason    string
	Version            string
	UpdatedAt          string
	PeggedOrder        PeggedOrder
	LiquidityProvision LiquidityProvision
}

type PeggedOrder struct {
	Reference string
	Offset    string
}

type LiquidityProvision struct {
	Id               string
	Party            Party
	CreatedAt        TimeString
	UpdatedAt        TimeString
	Market           Market
	CommitmentAmount string
	Fee              string
	// TODO: sells
	// TODO: buys
	Version   string
	Status    string
	Reference string
}

type Account struct {
	Balance string
	Asset   *Asset
	Type    string
	Market  *Market
}

type Asset struct {
	Id          string
	Name        string
	Symbol      string
	TotalSupply string
	Decimals    int
	Quantum     string
	// TODO: source
	InfrastructureFeeAccount Account
	GlobalRewardPoolAccount  Account
}

// ----------------------------------------------------------------------------
// Some wrappers around standard types to provide non-standard comparisons,
// where the output from the API might differ slightly but we don't care.

type HexString string

func (s HexString) Equal(other HexString) bool {
	// Don't care about casing of hex strings
	return strings.EqualFold(string(s), string(other))
}

type TimeString string

func (s TimeString) Equal(other TimeString) bool {
	if s == "" && other == "" {
		return true
	}
	// Postgres doesn't store nanoseconds, so only compare the millisecond portion
	t1, err1 := time.Parse(time.RFC3339Nano, string(s))
	t2, err2 := time.Parse(time.RFC3339Nano, string(other))

	if err1 != nil || err2 != nil {
		return false
	}

	t1t := t1.Truncate(time.Microsecond)
	t2t := t2.Truncate(time.Microsecond)
	if t1t != t2t {
		_ = "foo"
	}
	return t1t == t2t
}
