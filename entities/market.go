package entities

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	"code.vegaprotocol.io/protos/vega"
	"google.golang.org/protobuf/encoding/protojson"
)

type Market struct {
	ID                            []byte
	VegaTime                      time.Time
	InstrumentID                  string
	TradableInstrument            []byte
	DecimalPlaces                 int
	Fees                          []byte
	OpeningAuction                []byte
	PriceMonitoringSettings       []byte
	LiquidityMonitoringParameters []byte
	TradingMode                   MarketTradingMode
	State                         MarketState
	MarketTimestamps              []byte
	PositionDecimalPlaces         int
}

func MakeMarketID(stringID string) ([]byte, error) {
	id, err := hex.DecodeString(stringID)
	if err != nil {
		return nil, fmt.Errorf("market id is not valid hex string: %v", stringID)
	}
	return id, nil
}

func (m Market) HexID() string {
	return hex.EncodeToString(m.ID)
}

func NewMarketFromProto(market *vega.Market, vegaTime time.Time) (*Market, error) {
	id, err := MakeMarketID(market.Id)

	if err != nil {
		return nil, err
	}

	var tradableInstrument, fees, openingAuction, priceMonitoringSettings, liquidityMonitoringParameters, marketTimestamps []byte

	if market.TradableInstrument == nil {
		return nil, errors.New("tradable instrument cannot be nil")
	}

	if tradableInstrument, err = protojson.Marshal(market.TradableInstrument); err != nil {
		return nil, err
	}

	if market.Fees == nil {
		return nil, errors.New("fees cannot be nil")
	}

	if fees, err = protojson.Marshal(market.Fees); err != nil {
		return nil, err
	}

	if market.OpeningAuction != nil {
		if openingAuction, err = protojson.Marshal(market.OpeningAuction); err != nil {
			return nil, err
		}
	}

	if market.PriceMonitoringSettings == nil {
		return nil, errors.New("price monitoring settings cannot be nil")
	}

	if priceMonitoringSettings, err = protojson.Marshal(market.PriceMonitoringSettings); err != nil {
		return nil, err
	}

	if market.LiquidityMonitoringParameters == nil {
		return nil, errors.New("liquidity monitoring parameters cannot be nil")
	}

	if liquidityMonitoringParameters, err = protojson.Marshal(market.LiquidityMonitoringParameters); err != nil {
		return nil, err
	}

	if market.MarketTimestamps == nil {
		return nil, errors.New("market timestamps cannot be nil")
	}
	if marketTimestamps, err = protojson.Marshal(market.MarketTimestamps); err != nil {
		return nil, err
	}

	if market.DecimalPlaces > math.MaxInt {
		return nil, fmt.Errorf("%d is not a valid number for decimal places", market.DecimalPlaces)
	}

	if market.PositionDecimalPlaces > math.MaxInt {
		return nil, fmt.Errorf("%d is not a valid number for position decimal places", market.PositionDecimalPlaces)
	}

	dps := int(market.DecimalPlaces)
	positionDps := int(market.PositionDecimalPlaces)

	return &Market{
		ID:                            id,
		VegaTime:                      vegaTime,
		InstrumentID:                  market.TradableInstrument.Instrument.Id,
		TradableInstrument:            tradableInstrument,
		DecimalPlaces:                 dps,
		Fees:                          fees,
		OpeningAuction:                openingAuction,
		PriceMonitoringSettings:       priceMonitoringSettings,
		LiquidityMonitoringParameters: liquidityMonitoringParameters,
		TradingMode:                   MarketTradingMode(market.TradingMode),
		State:                         MarketState(market.State),
		MarketTimestamps:              marketTimestamps,
		PositionDecimalPlaces:         positionDps,
	}, nil
}

func (m Market) ToProto() (*vega.Market, error) {
	var tradableInstrument vega.TradableInstrument
	var fees vega.Fees
	var openingAuction vega.AuctionDuration
	var priceMonitoringSettings vega.PriceMonitoringSettings
	var liquidityMonitoringParameters vega.LiquidityMonitoringParameters
	var marketTimestamps vega.MarketTimestamps

	if err := protojson.Unmarshal(m.TradableInstrument, &tradableInstrument); err != nil {
		return nil, err
	}

	if err := protojson.Unmarshal(m.Fees, &fees); err != nil {
		return nil, err
	}

	if err := protojson.Unmarshal(m.OpeningAuction, &openingAuction); err != nil {
		return nil, err
	}

	if err := protojson.Unmarshal(m.PriceMonitoringSettings, &priceMonitoringSettings); err != nil {
		return nil, err
	}

	if err := protojson.Unmarshal(m.LiquidityMonitoringParameters, &liquidityMonitoringParameters); err != nil {
		return nil, err
	}

	if err := protojson.Unmarshal(m.MarketTimestamps, &marketTimestamps); err != nil {
		return nil, err
	}

	return &vega.Market{
		Id:                            m.HexID(),
		TradableInstrument:            &tradableInstrument,
		DecimalPlaces:                 uint64(m.DecimalPlaces),
		Fees:                          &fees,
		OpeningAuction:                &openingAuction,
		PriceMonitoringSettings:       &priceMonitoringSettings,
		LiquidityMonitoringParameters: &liquidityMonitoringParameters,
		TradingMode:                   vega.Market_TradingMode(m.TradingMode),
		State:                         vega.Market_State(m.State),
		MarketTimestamps:              &marketTimestamps,
		PositionDecimalPlaces:         uint64(m.PositionDecimalPlaces),
	}, nil
}
