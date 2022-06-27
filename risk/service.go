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

package risk

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"code.vegaprotocol.io/data-node/contextutil"
	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/data-node/storage"
	"code.vegaprotocol.io/vega/types/num"

	ptypes "code.vegaprotocol.io/protos/vega"
)

var ErrInvalidOrderSide = errors.New("invalid order side")

// RiskStore ...
//go:generate go run github.com/golang/mock/mockgen -destination mocks/risk_store_mock.go -package mocks code.vegaprotocol.io/data-node/risk RiskStore
type RiskStore interface {
	GetMarginLevelsByID(partyID string, marketID string) ([]ptypes.MarginLevels, error)
	GetMarketRiskFactors(marketID string) (ptypes.RiskFactor, error)
	Subscribe(c chan []ptypes.MarginLevels) uint64
	Unsubscribe(id uint64) error
}

// MarketDataStore ...
//go:generate go run github.com/golang/mock/mockgen -destination mocks/market_data_store_mock.go -package mocks code.vegaprotocol.io/data-node/risk MarketDataStore
type MarketDataStore interface {
	GetByID(string) (ptypes.MarketData, error)
}

// MarketStore ...
//go:generate go run github.com/golang/mock/mockgen -destination mocks/market_store_mock.go -package mocks code.vegaprotocol.io/data-node/risk MarketStore
type MarketStore interface {
	GetByID(string) (*ptypes.Market, error)
}

// Svc represent the market service
type Svc struct {
	Config
	log           *logging.Logger
	store         RiskStore
	mktDataStore  MarketDataStore
	mktStore      MarketStore
	subscriberCnt int32
}

func NewService(
	log *logging.Logger,
	c Config,
	store RiskStore,
	mktStore MarketStore,
	mktDataStore MarketDataStore,
) *Svc {
	log = log.Named(namedLogger)
	log.SetLevel(c.Level.Get())

	return &Svc{
		Config:       c,
		log:          log,
		store:        store,
		mktStore:     mktStore,
		mktDataStore: mktDataStore,
	}
}

// ReloadConf update the risk service internal configuration
func (s *Svc) ReloadConf(cfg Config) {
	s.log.Info("reloading configuration")
	if s.log.GetLevel() != cfg.Level.Get() {
		s.log.Info("updating log level",
			logging.String("old", s.log.GetLevel().String()),
			logging.String("new", cfg.Level.String()),
		)
		s.log.SetLevel(cfg.Level.Get())
	}

	s.Config = cfg
}

// GetMarginLevelsByID returns the margin levels for a given party
func (s *Svc) GetMarginLevelsByID(partyID, marketID string) ([]ptypes.MarginLevels, error) {
	marginLevels, err := s.store.GetMarginLevelsByID(partyID, marketID)
	// Searching for margin-levels by party, should return without error in this case
	// as just because a party has not traded does not mean they don't exist in vega
	if err != nil && (err != storage.ErrNoMarginLevelsForParty && err != storage.ErrNoMarginLevelsForMarket) {
		return nil, err
	} else {
		return marginLevels, nil
	}
}

func (s *Svc) EstimateMargin(ctx context.Context, order *ptypes.Order) (*ptypes.MarginLevels, error) {
	// first get the risk factors and market data (marketdata->markprice)
	rf, err := s.store.GetMarketRiskFactors(order.MarketId)
	if err != nil {
		return nil, err
	}
	mkt, err := s.mktStore.GetByID(order.MarketId)
	if err != nil {
		return nil, err
	}
	mktData, err := s.mktDataStore.GetByID(order.MarketId)
	if err != nil {
		return nil, err
	}

	if order.Side == ptypes.Side_SIDE_UNSPECIFIED {
		return nil, ErrInvalidOrderSide
	}

	f, err := num.DecimalFromString(rf.Short)
	if err != nil {
		return nil, err
	}
	if order.Side == ptypes.Side_SIDE_BUY {
		f, err = num.DecimalFromString(rf.Long)
		if err != nil {
			return nil, err
		}
	}

	asset, err := mkt.GetAsset()
	if err != nil {
		return nil, err
	}

	// now calculate margin maintenance
	markPrice, _ := num.DecimalFromString(mktData.MarkPrice)

	// if the order is a limit order, use the limit price to calculate the margin maintenance
	if order.Type == ptypes.Order_TYPE_MARKET {
		markPrice, _ = num.DecimalFromString(order.Price)
	}

	maintenanceMargin := num.DecimalFromFloat(float64(order.Size)).Mul(f).Mul(markPrice)
	// now we use the risk factors
	return &ptypes.MarginLevels{
		PartyId:                order.PartyId,
		MarketId:               mkt.GetId(),
		Asset:                  asset,
		Timestamp:              0,
		MaintenanceMargin:      maintenanceMargin.String(),
		SearchLevel:            maintenanceMargin.Mul(num.DecimalFromFloat(mkt.TradableInstrument.MarginCalculator.ScalingFactors.SearchLevel)).String(),
		InitialMargin:          maintenanceMargin.Mul(num.DecimalFromFloat(mkt.TradableInstrument.MarginCalculator.ScalingFactors.InitialMargin)).String(),
		CollateralReleaseLevel: maintenanceMargin.Mul(num.DecimalFromFloat(mkt.TradableInstrument.MarginCalculator.ScalingFactors.CollateralRelease)).String(),
	}, nil
}

func (s *Svc) ObserveMarginLevels(
	ctx context.Context, retries int, partyID, marketID string,
) (accountCh <-chan []ptypes.MarginLevels, ref uint64) {
	margins := make(chan []ptypes.MarginLevels)
	internal := make(chan []ptypes.MarginLevels)
	ref = s.store.Subscribe(internal)

	var cancel func()
	ctx, cancel = context.WithCancel(ctx)
	go func() {
		atomic.AddInt32(&s.subscriberCnt, 1)
		defer atomic.AddInt32(&s.subscriberCnt, -1)
		ip, _ := contextutil.RemoteIPAddrFromContext(ctx)
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				s.log.Debug(
					"Risks subscriber closed connection",
					logging.Uint64("id", ref),
					logging.String("ip-address", ip),
				)
				// this error only happens when the subscriber reference doesn't exist
				// so we can still safely close the channels
				if err := s.store.Unsubscribe(ref); err != nil {
					s.log.Error(
						"Failure un-subscribing accounts subscriber when context.Done()",
						logging.Uint64("id", ref),
						logging.String("ip-address", ip),
						logging.Error(err),
					)
				}
				close(internal)
				close(margins)
				return
			case accs := <-internal:
				filtered := make([]ptypes.MarginLevels, 0, len(accs))
				for _, acc := range accs {
					acc := acc
					if (len(marketID) <= 0 || marketID == acc.MarketId) &&
						partyID == acc.PartyId {
						filtered = append(filtered, acc)
					}
				}
				retryCount := retries
				success := false
				for !success && retryCount >= 0 {
					select {
					case margins <- filtered:
						retryCount = retries
						s.log.Debug(
							"Risks for subscriber sent successfully",
							logging.Uint64("ref", ref),
							logging.String("ip-address", ip),
						)
						success = true
					default:
						retryCount--
						if retryCount > 0 {
							s.log.Debug(
								"Risks for subscriber not sent",
								logging.Uint64("ref", ref),
								logging.String("ip-address", ip))
						}
						time.Sleep(time.Duration(10) * time.Millisecond)
					}
				}
				if !success && retryCount <= 0 {
					s.log.Warn(
						"Risk subscriber has hit the retry limit",
						logging.Uint64("ref", ref),
						logging.String("ip-address", ip),
						logging.Int("retries", retries))
					cancel()
				}

			}
		}
	}()

	return margins, ref
}

// GetMarginLevelsSubscribersCount returns the total number of active subscribers for ObserveMarginLevels.
func (s *Svc) GetMarginLevelsSubscribersCount() int32 {
	return atomic.LoadInt32(&s.subscriberCnt)
}

func (s *Svc) GetMarketRiskFactors(marketID string) (ptypes.RiskFactor, error) {
	return s.store.GetMarketRiskFactors(marketID)
}
