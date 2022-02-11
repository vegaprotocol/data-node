package entities

import (
	"bytes"
	"encoding/hex"
	"time"

	types "code.vegaprotocol.io/protos/vega"
	"github.com/shopspring/decimal"
)

// MarketData represents a market data record that is stored in the SQL database
type MarketData struct {
	// Mark price, as an integer, for example `123456` is a correctly
	// formatted price of `1.23456` assuming market configured to 5 decimal places
	MarkPrice *decimal.Decimal
	// Highest price level on an order book for buy orders, as an integer, for example `123456` is a correctly
	// formatted price of `1.23456` assuming market configured to 5 decimal places
	BestBidPrice *decimal.Decimal
	// Aggregated volume being bid at the best bid price
	BestBidVolume uint64
	// Aggregated volume being bid at the best bid price
	BestOfferPrice *decimal.Decimal
	// Aggregated volume being offered at the best offer price, as an integer, for example `123456` is a correctly
	// formatted price of `1.23456` assuming market configured to 5 decimal places
	BestOfferVolume uint64
	// Highest price on the order book for buy orders not including pegged orders
	BestStaticBidPrice *decimal.Decimal
	// Total volume at the best static bid price excluding pegged orders
	BestStaticBidVolume uint64
	// Lowest price on the order book for sell orders not including pegged orders
	BestStaticOfferPrice *decimal.Decimal
	// Total volume at the best static offer price excluding pegged orders
	BestStaticOfferVolume uint64
	// Arithmetic average of the best bid price and best offer price, as an integer, for example `123456` is a correctly
	// formatted price of `1.23456` assuming market configured to 5 decimal places
	MidPrice *decimal.Decimal
	// Arithmetic average of the best static bid price and best static offer price
	StaticMidPrice *decimal.Decimal
	// Market identifier for the data
	Market []byte
	// Timestamp at which this mark price was relevant, in nanoseconds since the epoch
	// - See [`VegaTimeResponse`](#api.VegaTimeResponse).`timestamp`
	Timestamp time.Time
	// The sum of the size of all positions greater than 0 on the market
	OpenInterest uint64
	// Time in seconds until the end of the auction (0 if currently not in auction period)
	AuctionEnd int64
	// Time until next auction (used in FBA's) - currently always 0
	AuctionStart int64
	// Indicative price (zero if not in auction)
	IndicativePrice *decimal.Decimal
	// Indicative volume (zero if not in auction)
	IndicativeVolume uint64
	// The current trading mode for the market
	MarketTradingMode string
	// When a market is in an auction trading mode, this field indicates what triggered the auction
	AuctionTrigger string
	// When a market auction is extended, this field indicates what caused the extension
	ExtensionTrigger string
	// Targeted stake for the given market
	TargetStake *decimal.Decimal
	// Available stake for the given market
	SuppliedStake *decimal.Decimal
	// One or more price monitoring bounds for the current timestamp
	PriceMonitoringBounds []*PriceMonitoringBound
	// the market value proxy
	MarketValueProxy string
	// the equity like share of liquidity fee for each liquidity provider
	LiquidityProviderFeeShares []*LiquidityProviderFeeShare
	// Vega Block time at which the data was received from Vega Node
	VegaTime time.Time
}

type PriceMonitoringTrigger struct {
	Horizon          int64   `json:"horizon,omitempty"`
	Probability      float64 `json:"probability,omitempty"`
	AuctionExtension int64   `json:"auctionExtension,omitempty"`
}

func (trigger PriceMonitoringTrigger) Equals(other PriceMonitoringTrigger) bool {
	return trigger.Horizon == other.Horizon &&
		trigger.Probability == other.Probability &&
		trigger.AuctionExtension == other.AuctionExtension
}

type PriceMonitoringBound struct {
	MinValidPrice string                 `json:"minValidPrice,omitempty"`
	MaxValidPrice string                 `json:"maxValidPrice,omitempty"`
	Trigger       PriceMonitoringTrigger `json:"trigger,omitempty"`
}

func (bound PriceMonitoringBound) Equals(other PriceMonitoringBound) bool {
	if bound.MinValidPrice != other.MinValidPrice {
		return false
	}

	if bound.MaxValidPrice != other.MaxValidPrice {
		return false
	}

	return bound.Trigger.Equals(other.Trigger)
}

type LiquidityProviderFeeShare struct {
	Party                 string `json:"party,omitempty"`
	EquityLikeShare       string `json:"equityLikeShare,omitempty"`
	AverageEntryValuation string `json:"averageEntryValuation,omitempty"`
}

func (fee LiquidityProviderFeeShare) Equals(other LiquidityProviderFeeShare) bool {
	return fee.Party == other.Party &&
		fee.EquityLikeShare == other.EquityLikeShare &&
		fee.AverageEntryValuation == other.AverageEntryValuation
}

func MarketDataFromProto(data types.MarketData) (*MarketData, error) {
	var mark, bid, offer, staticBid, staticOffer, mid, staticMid, indicative, targetStake, suppliedStake *decimal.Decimal
	var err error
	var marketID []byte

	if marketID, err = hex.DecodeString(data.Market); err != nil {
		return nil, nil
	}
	if mark, err = parseDecimal(data.MarkPrice); err != nil {
		return nil, err
	}
	if bid, err = parseDecimal(data.BestBidPrice); err != nil {
		return nil, err
	}
	if offer, err = parseDecimal(data.BestOfferPrice); err != nil {
		return nil, err
	}
	if staticBid, err = parseDecimal(data.BestStaticBidPrice); err != nil {
		return nil, err
	}
	if staticOffer, err = parseDecimal(data.BestStaticOfferPrice); err != nil {
		return nil, err
	}
	if mid, err = parseDecimal(data.MidPrice); err != nil {
		return nil, err
	}
	if staticMid, err = parseDecimal(data.StaticMidPrice); err != nil {
		return nil, err
	}
	if indicative, err = parseDecimal(data.IndicativePrice); err != nil {
		return nil, err
	}
	if targetStake, err = parseDecimal(data.TargetStake); err != nil {
		return nil, err
	}
	if suppliedStake, err = parseDecimal(data.SuppliedStake); err != nil {
		return nil, err
	}
	ts := time.Unix(0, data.Timestamp).UTC()

	marketData := &MarketData{
		MarkPrice:                  mark,
		BestBidPrice:               bid,
		BestBidVolume:              data.BestBidVolume,
		BestOfferPrice:             offer,
		BestOfferVolume:            data.BestOfferVolume,
		BestStaticBidPrice:         staticBid,
		BestStaticBidVolume:        data.BestStaticBidVolume,
		BestStaticOfferPrice:       staticOffer,
		BestStaticOfferVolume:      data.BestStaticOfferVolume,
		MidPrice:                   mid,
		StaticMidPrice:             staticMid,
		Market:                     marketID,
		Timestamp:                  ts,
		OpenInterest:               data.OpenInterest,
		AuctionEnd:                 data.AuctionEnd,
		AuctionStart:               data.AuctionStart,
		IndicativePrice:            indicative,
		IndicativeVolume:           data.IndicativeVolume,
		MarketTradingMode:          data.MarketTradingMode.String(),
		AuctionTrigger:             data.Trigger.String(),
		ExtensionTrigger:           data.ExtensionTrigger.String(),
		TargetStake:                targetStake,
		SuppliedStake:              suppliedStake,
		PriceMonitoringBounds:      parsePriceMonitoringBounds(data.PriceMonitoringBounds),
		MarketValueProxy:           data.MarketValueProxy,
		LiquidityProviderFeeShares: parseLiquidityProviderFeeShares(data.LiquidityProviderFeeShare),
	}

	return marketData, nil
}

func parseDecimal(input string) (*decimal.Decimal, error) {
	if input == "" {
		return nil, nil
	}

	v, err := decimal.NewFromString(input)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func parsePriceMonitoringBounds(bounds []*types.PriceMonitoringBounds) []*PriceMonitoringBound {
	if len(bounds) == 0 {
		return nil
	}

	results := make([]*PriceMonitoringBound, 0, len(bounds))

	for _, b := range bounds {
		results = append(results, priceMonitoringBoundsFromProto(b))
	}

	return results
}

func parseLiquidityProviderFeeShares(shares []*types.LiquidityProviderFeeShare) []*LiquidityProviderFeeShare {
	if len(shares) == 0 {
		return nil
	}

	results := make([]*LiquidityProviderFeeShare, 0, len(shares))

	for _, s := range shares {
		results = append(results, liquidityProviderFeeShareFromProto(s))
	}

	return results
}

func priceMonitoringBoundsFromProto(bounds *types.PriceMonitoringBounds) *PriceMonitoringBound {
	if bounds == nil {
		return nil
	}

	return &PriceMonitoringBound{
		MinValidPrice: bounds.MinValidPrice,
		MaxValidPrice: bounds.MaxValidPrice,
		Trigger:       priceMonitoringTriggerFromProto(bounds.Trigger),
	}
}

func priceMonitoringTriggerFromProto(trigger *types.PriceMonitoringTrigger) PriceMonitoringTrigger {
	if trigger == nil {
		return PriceMonitoringTrigger{}
	}

	return PriceMonitoringTrigger{
		Horizon:          trigger.Horizon,
		Probability:      trigger.Probability,
		AuctionExtension: trigger.AuctionExtension,
	}
}

func liquidityProviderFeeShareFromProto(feeShare *types.LiquidityProviderFeeShare) *LiquidityProviderFeeShare {
	if feeShare == nil {
		return nil
	}

	return &LiquidityProviderFeeShare{
		Party:                 feeShare.Party,
		EquityLikeShare:       feeShare.EquityLikeShare,
		AverageEntryValuation: feeShare.AverageEntryValuation,
	}
}

func decimalRefIsEqual(value, other *decimal.Decimal) bool {
	if value == nil && other == nil {
		return true
	}
	if value != nil && other != nil {
		return value.Equal(*other)
	}
	return false
}

func (md MarketData) Equal(other MarketData) bool {
	return decimalRefIsEqual(md.MarkPrice, other.MarkPrice) &&
		decimalRefIsEqual(md.BestBidPrice, other.BestBidPrice) &&
		decimalRefIsEqual(md.BestOfferPrice, other.BestOfferPrice) &&
		decimalRefIsEqual(md.BestStaticBidPrice, other.BestStaticBidPrice) &&
		decimalRefIsEqual(md.BestStaticOfferPrice, other.BestStaticOfferPrice) &&
		decimalRefIsEqual(md.MidPrice, other.MidPrice) &&
		decimalRefIsEqual(md.StaticMidPrice, other.StaticMidPrice) &&
		decimalRefIsEqual(md.IndicativePrice, other.IndicativePrice) &&
		decimalRefIsEqual(md.TargetStake, other.TargetStake) &&
		decimalRefIsEqual(md.SuppliedStake, other.SuppliedStake) &&
		md.BestBidVolume == other.BestBidVolume &&
		md.BestOfferVolume == other.BestOfferVolume &&
		md.BestStaticBidVolume == other.BestStaticBidVolume &&
		md.BestStaticOfferVolume == other.BestStaticOfferVolume &&
		md.OpenInterest == other.OpenInterest &&
		md.AuctionEnd == other.AuctionEnd &&
		md.AuctionStart == other.AuctionStart &&
		md.IndicativeVolume == other.IndicativeVolume &&
		bytes.Equal(md.Market, other.Market) &&
		md.MarketTradingMode == other.MarketTradingMode &&
		md.AuctionTrigger == other.AuctionTrigger &&
		md.ExtensionTrigger == other.ExtensionTrigger &&
		md.MarketValueProxy == other.MarketValueProxy &&
		md.Timestamp.Equal(other.Timestamp) &&
		priceMonitoringBoundsMatches(md.PriceMonitoringBounds, other.PriceMonitoringBounds) &&
		liquidityProviderFeeShareMatches(md.LiquidityProviderFeeShares, other.LiquidityProviderFeeShares)
}

func priceMonitoringBoundsMatches(bounds, other []*PriceMonitoringBound) bool {
	if len(bounds) != len(other) {
		return false
	}

	for i, bound := range bounds {
		if !bound.Equals(*other[i]) {
			return false
		}
	}

	return true
}

func liquidityProviderFeeShareMatches(feeShares, other []*LiquidityProviderFeeShare) bool {
	if len(feeShares) != len(other) {
		return false
	}

	for i, fee := range feeShares {
		if !fee.Equals(*other[i]) {
			return false
		}
	}

	return true
}
