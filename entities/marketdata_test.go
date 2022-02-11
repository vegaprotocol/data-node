package entities_test

import (
	"testing"
	"time"

	"code.vegaprotocol.io/data-node/entities"
	types "code.vegaprotocol.io/protos/vega"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestMarketDataFromProto(t *testing.T) {
	t.Run("should parse all valid prices", testParseAllValidPrices)
	t.Run("should return nil prices if string is empty", testParseEmptyPrices)
	t.Run("should return error if an invalid price string is provided", testParseInvalidPriceString)
	t.Run("should parse valid market data records successfully", testParseMarketDataSuccessfully)
}

func testParseAllValidPrices(t *testing.T) {
	marketdata := types.MarketData{
		MarkPrice:            "1",
		BestBidPrice:         "1",
		BestOfferPrice:       "1",
		BestStaticBidPrice:   "1",
		BestStaticOfferPrice: "1",
		MidPrice:             "1",
		StaticMidPrice:       "1",
		IndicativePrice:      "1",
		TargetStake:          "1",
		SuppliedStake:        "1",
	}

	md, err := entities.MarketDataFromProto(marketdata)
	assert.NoError(t, err)
	assert.NotNil(t, md.MarkPrice)
	assert.NotNil(t, md.BestBidPrice)
	assert.NotNil(t, md.BestOfferPrice)
	assert.NotNil(t, md.BestStaticBidPrice)
	assert.NotNil(t, md.BestStaticOfferVolume)
	assert.NotNil(t, md.MidPrice)
	assert.NotNil(t, md.StaticMidPrice)
	assert.NotNil(t, md.IndicativePrice)
	assert.NotNil(t, md.TargetStake)
	assert.NotNil(t, md.SuppliedStake)

	want := decimal.NewFromInt(1)
	assert.True(t, want.Equal(*md.MarkPrice))
	assert.True(t, want.Equal(*md.BestBidPrice))
	assert.True(t, want.Equal(*md.BestOfferPrice))
	assert.True(t, want.Equal(*md.BestStaticBidPrice))
	assert.True(t, want.Equal(*md.BestStaticOfferPrice))
	assert.True(t, want.Equal(*md.MidPrice))
	assert.True(t, want.Equal(*md.StaticMidPrice))
	assert.True(t, want.Equal(*md.IndicativePrice))
	assert.True(t, want.Equal(*md.TargetStake))
	assert.True(t, want.Equal(*md.SuppliedStake))
}

func testParseEmptyPrices(t *testing.T) {
	marketdata := types.MarketData{}
	md, err := entities.MarketDataFromProto(marketdata)
	assert.NoError(t, err)
	assert.Nil(t, md.MarkPrice)
	assert.Nil(t, md.BestBidPrice)
	assert.Nil(t, md.BestOfferPrice)
	assert.Nil(t, md.BestStaticBidPrice)
	assert.Nil(t, md.BestStaticOfferPrice)
	assert.Nil(t, md.MidPrice)
	assert.Nil(t, md.StaticMidPrice)
	assert.Nil(t, md.IndicativePrice)
	assert.Nil(t, md.TargetStake)
	assert.Nil(t, md.SuppliedStake)
}

func testParseInvalidPriceString(t *testing.T) {
	type args struct {
		marketdata types.MarketData
	}
	testCases := []struct {
		name string
		args args
	}{
		{
			name: "Invalid Mark Price",
			args: args{
				marketdata: types.MarketData{
					MarkPrice: "Test",
				},
			},
		},
		{
			name: "Invalid Best Bid Price",
			args: args{
				marketdata: types.MarketData{
					BestBidPrice: "Test",
				},
			},
		},
		{
			name: "Invalid Best Offer Price",
			args: args{
				marketdata: types.MarketData{
					BestOfferPrice: "Test",
				},
			},
		},
		{
			name: "Invalid Best Static Bid Price",
			args: args{
				marketdata: types.MarketData{
					BestStaticBidPrice: "Test",
				},
			},
		},
		{
			name: "Invalid Best Static Offer Price",
			args: args{
				marketdata: types.MarketData{
					BestStaticOfferPrice: "Test",
				},
			},
		},
		{
			name: "Invalid Mid Price",
			args: args{
				marketdata: types.MarketData{
					MidPrice: "Test",
				},
			},
		},
		{
			name: "Invalid Static Mid Price",
			args: args{
				marketdata: types.MarketData{
					StaticMidPrice: "Test",
				},
			},
		},
		{
			name: "Invalid Indicative Price",
			args: args{
				marketdata: types.MarketData{
					IndicativePrice: "Test",
				},
			},
		},
		{
			name: "Invalid Target Stake",
			args: args{
				marketdata: types.MarketData{
					TargetStake: "Test",
				},
			},
		},
		{
			name: "Invalid Supplied Stake",
			args: args{
				marketdata: types.MarketData{
					SuppliedStake: "Test",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			md, err := entities.MarketDataFromProto(tc.args.marketdata)
			assert.Error(tt, err)
			assert.Nil(tt, md)
		})
	}
}

func testParseMarketDataSuccessfully(t *testing.T) {
	type args struct {
		marketdata types.MarketData
	}

	testCases := []struct {
		name string
		args args
		want *entities.MarketData
	}{
		{
			name: "Empty market data",
			args: args{
				marketdata: types.MarketData{},
			},
			want: &entities.MarketData{
				Timestamp:         time.Unix(0, 0).UTC(),
				AuctionTrigger:    "AUCTION_TRIGGER_UNSPECIFIED",
				MarketTradingMode: "TRADING_MODE_UNSPECIFIED",
				ExtensionTrigger:  "AUCTION_TRIGGER_UNSPECIFIED",
			},
		},
		{
			name: "Market data with auction trigger specified",
			args: args{
				marketdata: types.MarketData{
					Trigger: types.AuctionTrigger_AUCTION_TRIGGER_PRICE,
				},
			},
			want: &entities.MarketData{
				Timestamp:         time.Unix(0, 0).UTC(),
				AuctionTrigger:    "AUCTION_TRIGGER_PRICE",
				MarketTradingMode: "TRADING_MODE_UNSPECIFIED",
				ExtensionTrigger:  "AUCTION_TRIGGER_UNSPECIFIED",
			},
		},
		{
			name: "Market data with auction trigger and market trading mode specified",
			args: args{
				marketdata: types.MarketData{
					Trigger:           types.AuctionTrigger_AUCTION_TRIGGER_PRICE,
					MarketTradingMode: types.Market_TRADING_MODE_CONTINUOUS,
				},
			},
			want: &entities.MarketData{
				Timestamp:         time.Unix(0, 0).UTC(),
				AuctionTrigger:    "AUCTION_TRIGGER_PRICE",
				MarketTradingMode: "TRADING_MODE_CONTINUOUS",
				ExtensionTrigger:  "AUCTION_TRIGGER_UNSPECIFIED",
			},
		},
		{
			name: "Market data with best bid and best offer specified",
			args: args{
				marketdata: types.MarketData{
					BestBidPrice:      "100.0",
					BestOfferPrice:    "110.0",
					Trigger:           types.AuctionTrigger_AUCTION_TRIGGER_PRICE,
					MarketTradingMode: types.Market_TRADING_MODE_CONTINUOUS,
				},
			},
			want: &entities.MarketData{
				BestBidPrice:      getDecimalRef(100.0),
				BestOfferPrice:    getDecimalRef(110.0),
				Timestamp:         time.Unix(0, 0).UTC(),
				AuctionTrigger:    "AUCTION_TRIGGER_PRICE",
				MarketTradingMode: "TRADING_MODE_CONTINUOUS",
				ExtensionTrigger:  "AUCTION_TRIGGER_UNSPECIFIED",
			},
		},
		{
			name: "Market data with best bid and best offer specified and price monitoring bounds",
			args: args{
				marketdata: types.MarketData{
					BestBidPrice:      "100.0",
					BestOfferPrice:    "110.0",
					Trigger:           types.AuctionTrigger_AUCTION_TRIGGER_PRICE,
					MarketTradingMode: types.Market_TRADING_MODE_CONTINUOUS,
					PriceMonitoringBounds: []*types.PriceMonitoringBounds{
						{
							MinValidPrice: "100",
							MaxValidPrice: "200",
						},
					},
				},
			},
			want: &entities.MarketData{
				BestBidPrice:      getDecimalRef(100.0),
				BestOfferPrice:    getDecimalRef(110.0),
				Timestamp:         time.Unix(0, 0).UTC(),
				AuctionTrigger:    "AUCTION_TRIGGER_PRICE",
				MarketTradingMode: "TRADING_MODE_CONTINUOUS",
				ExtensionTrigger:  "AUCTION_TRIGGER_UNSPECIFIED",
				PriceMonitoringBounds: []*entities.PriceMonitoringBound{
					{
						MinValidPrice: "100",
						MaxValidPrice: "200",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			got, err := entities.MarketDataFromProto(tc.args.marketdata)
			assert.NoError(tt, err)
			assert.True(tt, tc.want.Equal(*got))
		})
	}
}

func getDecimalRef(v float64) *decimal.Decimal {
	d := decimal.NewFromFloat(v)
	return &d
}
