package risk_test

import (
	"context"
	"errors"
	"testing"

	"code.vegaprotocol.io/data-node/logging"
	ptypes "code.vegaprotocol.io/protos/vega"
	"code.vegaprotocol.io/data-node/risk"
	"code.vegaprotocol.io/data-node/risk/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type testSvc struct {
	*risk.Svc
	ctrl         *gomock.Controller
	mktstore     *mocks.MockMarketStore
	mktdatastore *mocks.MockMarketDataStore
	store        *mocks.MockRiskStore
}

func getTestSvc(t *testing.T) *testSvc {
	ctrl := gomock.NewController(t)
	mktstore := mocks.NewMockMarketStore(ctrl)
	mktdatastore := mocks.NewMockMarketDataStore(ctrl)
	store := mocks.NewMockRiskStore(ctrl)
	svc := risk.NewService(logging.NewTestLogger(), risk.NewDefaultConfig(),
		store, mktstore, mktdatastore)
	return &testSvc{
		Svc:          svc,
		ctrl:         ctrl,
		mktstore:     mktstore,
		mktdatastore: mktdatastore,
		store:        store,
	}
}

func TestMarginEstimates(t *testing.T) {
	t.Run("margin estimates success", testMarginEstimateSuccess)
	t.Run("margin estimates errors", testMarginEstimateErrors)
}

func testMarginEstimateSuccess(t *testing.T) {
	svc := getTestSvc(t)
	defer svc.ctrl.Finish()

	order := &ptypes.Order{
		Size:  10,
		Side:  ptypes.Side_SIDE_BUY,
		Price: 10,
	}

	svc.store.EXPECT().GetMarketRiskFactors(gomock.Any()).Times(1).Return(
		ptypes.RiskFactor{
			Long:  0.5,
			Short: 0.5,
		},
		nil,
	)
	svc.mktstore.EXPECT().GetByID(gomock.Any()).Times(1).Return(
		&ptypes.Market{
			Id: "mktid",
			TradableInstrument: &ptypes.TradableInstrument{
				Instrument: &ptypes.Instrument{
					Product: &ptypes.Instrument_Future{
						Future: &ptypes.Future{
							SettlementAsset: "assetid",
						},
					},
				},
				MarginCalculator: &ptypes.MarginCalculator{
					ScalingFactors: &ptypes.ScalingFactors{
						SearchLevel:       1.1,
						InitialMargin:     1.2,
						CollateralRelease: 1.3,
					},
				},
			},
		},
		nil,
	)
	svc.mktdatastore.EXPECT().GetByID(gomock.Any()).Times(1).Return(
		ptypes.MarketData{
			MarkPrice: 200,
		},
		nil,
	)

	lvls, err := svc.EstimateMargin(context.Background(), order)
	assert.NoError(t, err)
	assert.NotNil(t, lvls)
	assert.Equal(t, 1000, int(lvls.MaintenanceMargin))
	assert.Equal(t, 1100, int(lvls.SearchLevel))
	assert.Equal(t, 1200, int(lvls.InitialMargin))
	assert.Equal(t, 1300, int(lvls.CollateralReleaseLevel))
}

func testMarginEstimateErrors(t *testing.T) {
	svc := getTestSvc(t)
	defer svc.ctrl.Finish()

	order := &ptypes.Order{
		Size:  10,
		Price: 10,
	}

	// first test no risk factors

	svc.store.EXPECT().GetMarketRiskFactors(gomock.Any()).Times(1).Return(
		ptypes.RiskFactor{},
		errors.New("no risk factors"),
	)

	lvls, err := svc.EstimateMargin(context.Background(), order)
	assert.EqualError(t, err, "no risk factors")
	assert.Nil(t, lvls)

	// then test not mkt

	svc.store.EXPECT().GetMarketRiskFactors(gomock.Any()).Times(1).Return(
		ptypes.RiskFactor{
			Long:  0.5,
			Short: 0.5,
		},
		nil,
	)
	svc.mktstore.EXPECT().GetByID(gomock.Any()).Times(1).Return(
		nil,
		errors.New("no market"),
	)

	lvls, err = svc.EstimateMargin(context.Background(), order)
	assert.EqualError(t, err, "no market")
	assert.Nil(t, lvls)

	// then no market data

	svc.store.EXPECT().GetMarketRiskFactors(gomock.Any()).Times(1).Return(
		ptypes.RiskFactor{
			Long:  0.5,
			Short: 0.5,
		},
		nil,
	)

	svc.mktstore.EXPECT().GetByID(gomock.Any()).Times(1).Return(
		&ptypes.Market{
			TradableInstrument: &ptypes.TradableInstrument{
				MarginCalculator: &ptypes.MarginCalculator{
					ScalingFactors: &ptypes.ScalingFactors{
						SearchLevel:       1.1,
						InitialMargin:     1.2,
						CollateralRelease: 1.3,
					},
				},
			},
		},
		nil,
	)
	svc.mktdatastore.EXPECT().GetByID(gomock.Any()).Times(1).Return(
		ptypes.MarketData{},
		errors.New("no market data"),
	)

	lvls, err = svc.EstimateMargin(context.Background(), order)
	assert.EqualError(t, err, "no market data")
	assert.Nil(t, lvls)

	// then order is invalid

	svc.store.EXPECT().GetMarketRiskFactors(gomock.Any()).Times(1).Return(
		ptypes.RiskFactor{
			Long:  0.5,
			Short: 0.5,
		},
		nil,
	)

	svc.mktstore.EXPECT().GetByID(gomock.Any()).Times(1).Return(
		&ptypes.Market{
			TradableInstrument: &ptypes.TradableInstrument{
				MarginCalculator: &ptypes.MarginCalculator{
					ScalingFactors: &ptypes.ScalingFactors{
						SearchLevel:       1.1,
						InitialMargin:     1.2,
						CollateralRelease: 1.3,
					},
				},
			},
		},
		nil,
	)
	svc.mktdatastore.EXPECT().GetByID(gomock.Any()).Times(1).Return(
		ptypes.MarketData{
			MarkPrice: 200,
		},
		nil,
	)

	lvls, err = svc.EstimateMargin(context.Background(), order)
	assert.EqualError(t, err, risk.ErrInvalidOrderSide.Error())
	assert.Nil(t, lvls)
}
