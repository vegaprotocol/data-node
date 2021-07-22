package trades

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"code.vegaprotocol.io/data-node/contextutil"
	"code.vegaprotocol.io/data-node/logging"
	ptypes "code.vegaprotocol.io/data-node/proto/vega"
	"code.vegaprotocol.io/data-node/storage"
	"code.vegaprotocol.io/data-node/types"
)

// TradeStore represents an abstraction over a trade storage
//go:generate go run github.com/golang/mock/mockgen -destination mocks/trade_store_mock.go -package mocks code.vegaprotocol.io/data-node/trades TradeStore
type TradeStore interface {
	GetByMarket(ctx context.Context, market string, skip, limit uint64, descending bool) ([]*ptypes.Trade, error)
	GetByMarketAndID(ctx context.Context, market string, id string) (*ptypes.Trade, error)
	GetByParty(ctx context.Context, party string, skip, limit uint64, descending bool, market *string) ([]*ptypes.Trade, error)
	GetByPartyAndID(ctx context.Context, party string, id string) (*ptypes.Trade, error)
	GetByOrderID(ctx context.Context, orderID string, skip, limit uint64, descending bool, market *string) ([]*ptypes.Trade, error)
	GetTradesBySideBuckets(ctx context.Context, party string) map[string]*storage.MarketBucket
	GetMarkPrice(ctx context.Context, market string) (uint64, error)
	Subscribe(trades chan<- []ptypes.Trade) uint64
	Unsubscribe(id uint64) error
}

//go:generate go run github.com/golang/mock/mockgen -destination mocks/positions_plugin_mock.go -package mocks code.vegaprotocol.io/data-node/trades PositionsPlugin
type PositionsPlugin interface {
	GetPositionsByMarket(market string) ([]*types.Position, error)
	GetPositionsByParty(party string) ([]*types.Position, error)
	GetPositionsByMarketAndParty(market, party string) (*types.Position, error)
	GetAllPositions() ([]*types.Position, error)
}

// Svc is the service handling trades
type Svc struct {
	Config
	log                     *logging.Logger
	tradeStore              TradeStore
	positions               PositionsPlugin
	positionsSubscribersCnt int32
	tradeSubscribersCnt     int32
}

// NewService instantiate a new Trades service
func NewService(log *logging.Logger, config Config, tradeStore TradeStore, posPlug PositionsPlugin) (*Svc, error) {
	// setup logger
	log = log.Named(namedLogger)
	log.SetLevel(config.Level.Get())

	return &Svc{
		log:        log,
		Config:     config,
		tradeStore: tradeStore,
		positions:  posPlug,
	}, nil
}

// ReloadConf update the internal configuration of the service
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

func (s *Svc) checkPagination(limit *uint64) error {
	if *limit == 0 {
		*limit = s.Config.PageSizeDefault
		// Do not return yet. The default may have been set to a number greater
		// than the maximum.
	}

	if *limit > s.Config.PageSizeMaximum {
		return fmt.Errorf("invalid pagination limit: %d is greater than %d", *limit, s.Config.PageSizeMaximum)
	}

	return nil
}

//GetByMarket returns a list of trades for a given market
func (s *Svc) GetByMarket(ctx context.Context, market string, skip, limit uint64, descending bool) (trades []*ptypes.Trade, err error) {
	if err = s.checkPagination(&limit); err != nil {
		return nil, err
	}

	trades, err = s.tradeStore.GetByMarket(ctx, market, skip, limit, descending)
	if err != nil {
		return nil, err
	}
	return trades, err
}

// GetByParty returns a list of trade for a given party
func (s *Svc) GetByParty(ctx context.Context, party string, skip, limit uint64, descending bool, market *string) (trades []*ptypes.Trade, err error) {
	if err = s.checkPagination(&limit); err != nil {
		return nil, err
	}

	trades, err = s.tradeStore.GetByParty(ctx, party, skip, limit, descending, market)
	if err != nil {
		return nil, err
	}
	return trades, err
}

// GetByMarketAndID return a single trade per its ID and the market it was created in
func (s *Svc) GetByMarketAndID(ctx context.Context, market string, id string) (trade *ptypes.Trade, err error) {
	trade, err = s.tradeStore.GetByMarketAndID(ctx, market, id)
	if err != nil {
		return &ptypes.Trade{}, err
	}
	return trade, err
}

// GetByPartyAndID returns a single trade, filter through a party ID and the trade ID
func (s *Svc) GetByPartyAndID(ctx context.Context, party string, id string) (trade *ptypes.Trade, err error) {
	trade, err = s.tradeStore.GetByPartyAndID(ctx, party, id)
	if err != nil {
		return &ptypes.Trade{}, err
	}
	return trade, err
}

// GetByOrderID return a list of trades filter by order ID (even the buy or sell side of the trade)
func (s *Svc) GetByOrderID(ctx context.Context, orderID string) (trades []*ptypes.Trade, err error) {
	trades, err = s.tradeStore.GetByOrderID(ctx, orderID, 0, 0, false, nil)
	if err != nil {
		return nil, err
	}
	return trades, err
}

// GetTradeSubscribersCount return the count of subscribers to the Trades updates
func (s *Svc) GetTradeSubscribersCount() int32 {
	return atomic.LoadInt32(&s.tradeSubscribersCnt)
}

// ObserveTrades return a channel to the caller through which it will receive notification
// on all trades happening in the system.
func (s *Svc) ObserveTrades(ctx context.Context, retries int, market *string, party *string) (<-chan []ptypes.Trade, uint64) {
	trades := make(chan []ptypes.Trade)
	internal := make(chan []ptypes.Trade)
	ref := s.tradeStore.Subscribe(internal)

	var cancel func()
	ctx, cancel = context.WithCancel(ctx)
	go func() {
		atomic.AddInt32(&s.tradeSubscribersCnt, 1)
		defer atomic.AddInt32(&s.tradeSubscribersCnt, -1)
		ip, _ := contextutil.RemoteIPAddrFromContext(ctx)
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				s.log.Debug(
					"Trades subscriber closed connection",
					logging.Uint64("id", ref),
					logging.String("ip-address", ip),
				)
				if err := s.tradeStore.Unsubscribe(ref); err != nil {
					s.log.Error(
						"Failure un-subscribing trades subscriber when context.Done()",
						logging.Uint64("id", ref),
						logging.String("ip-address", ip),
						logging.Error(err),
					)
				}
				close(internal)
				close(trades)
				return
			case v := <-internal:
				// max length of validated == length of data from channel
				validatedTrades := make([]ptypes.Trade, 0, len(v))
				for _, item := range v {
					// if market is nil or matches item market and party was nil, or matches seller or buyer
					if (market == nil || item.MarketId == *market) && (party == nil || item.Seller == *party || item.Buyer == *party) {
						validatedTrades = append(validatedTrades, item)
					}
				}
				if len(validatedTrades) == 0 {
					continue
				}
				retryCount := retries
				success := false
				for !success && retryCount >= 0 {
					select {
					case trades <- validatedTrades:
						s.log.Debug(
							"Trades for subscriber sent successfully",
							logging.Uint64("ref", ref),
							logging.String("ip-address", ip),
						)
						success = true
					default:
						retryCount--
						if retryCount >= 0 {
							s.log.Debug(
								"Trades for subscriber not sent",
								logging.Uint64("ref", ref),
								logging.String("ip-address", ip),
							)
							time.Sleep(time.Duration(10) * time.Millisecond)
						}
					}
				}
				if !success && retryCount <= 0 {
					s.log.Warn(
						"Trades subscriber has hit the retry limit",
						logging.Uint64("ref", ref),
						logging.String("ip-address", ip),
						logging.Int("retries", retries),
					)
					cancel()
				}

			}
		}
	}()

	return trades, ref
}

// GetPositionsSubscribersCount return the number of subscriber to the positions observer
func (s *Svc) GetPositionsSubscribersCount() int32 {
	return atomic.LoadInt32(&s.positionsSubscribersCnt)
}

// ObservePositions return a channel through which all positions are streamed to the caller
// when they get updated
func (s *Svc) ObservePositions(ctx context.Context, retries int, party, market string) (<-chan *ptypes.Position, uint64) {
	pBuf := 1
	// all parties, increase channel buffer by 100
	if party == "" {
		pBuf *= 100
	}
	// all markets, increase buffer by 100 (so all parties in all markets == 1k buffer)
	if market == "" {
		pBuf *= 10
	}
	positions := make(chan *ptypes.Position, pBuf)
	internal := make(chan []ptypes.Trade)
	ref := s.tradeStore.Subscribe(internal)

	var cancel func()
	ctx, cancel = context.WithCancel(ctx)
	go func() {
		atomic.AddInt32(&s.positionsSubscribersCnt, 1)
		defer atomic.AddInt32(&s.positionsSubscribersCnt, -1)
		ip, _ := contextutil.RemoteIPAddrFromContext(ctx)
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				s.log.Debug(
					"Positions subscriber closed connection",
					logging.Uint64("id", ref),
					logging.String("ip-address", ip),
				)
				if err := s.tradeStore.Unsubscribe(ref); err != nil {
					s.log.Error(
						"Failure un-subscribing positions subscriber when context.Done()",
						logging.Uint64("id", ref),
						logging.String("ip-address", ip),
						logging.Error(err),
					)
				}
				close(internal)
				close(positions)
				return
			case <-internal: // again, we're using this channel to detect state changes, the data itself isn't relevant
				mapOfMarketPositions, err := s.GetPositionsByParty(ctx, party, market)
				if err != nil {
					s.log.Error(
						"Failed to get positions for subscriber (getPositionsByParty)",
						logging.Uint64("ref", ref),
						logging.String("ip-address", ip),
						logging.Error(err),
					)
					continue
				}
				for _, marketPositions := range mapOfMarketPositions {
					marketPositions := marketPositions
					retryCount := retries
					success := false
					for !success && retryCount > 0 {
						select {
						case positions <- marketPositions:
							s.log.Debug(
								"Positions for subscriber sent successfully",
								logging.Uint64("ref", ref),
								logging.String("ip-address", ip),
							)
							success = true
						default:
							retryCount--
							if retryCount > 0 {
								s.log.Debug(
									"Positions for subscriber not sent",
									logging.Uint64("ref", ref),
									logging.String("ip-address", ip),
								)
								time.Sleep(time.Duration(10) * time.Millisecond)
							}
						}
					}
					if retryCount <= 0 {
						s.log.Warn(
							"Positions subscriber has hit the retry limit",
							logging.Uint64("ref", ref),
							logging.String("ip-address", ip),
							logging.Int("retries", retries),
						)
						cancel()
						break
					}

				}
			}
		}
	}()

	return positions, ref
}

// GetPositionsByParty returns a list of positions for a given party
func (s *Svc) GetPositionsByParty(ctx context.Context, party, marketID string) ([]*ptypes.Position, error) {

	s.log.Debug("Calculate positions for party",
		logging.String("party-id", party))

	if party == "" && marketID == "" {
		pos, err := s.positions.GetAllPositions()
		if err != nil {
			s.log.Error("Error getting all positions", logging.Error(err))
			return nil, err
		}
		return types.Positions(pos).IntoProto(), nil
	}
	var positions []*types.Position
	if party != "" && marketID != "" {
		pos, err := s.positions.GetPositionsByMarketAndParty(marketID, party)
		if err != nil {
			s.log.Error("Error getting position for party and market",
				logging.String("party-id", party),
				logging.String("market-id", marketID),
				logging.Error(err))
			return nil, err
		}
		if pos != nil {
			positions = []*types.Position{pos}
		}
	} else if party == "" {
		// either trader or market == ""
		pos, err := s.positions.GetPositionsByMarket(marketID)
		if err != nil {
			s.log.Error("Error getting positions for market",
				logging.String("market-id", marketID),
				logging.Error(err),
			)
			return nil, err
		}
		positions = pos
	} else {
		pos, err := s.positions.GetPositionsByParty(party)
		if err != nil {
			s.log.Error("Error getting positions for market",
				logging.String("party-id", party),
				logging.Error(err),
			)
			return nil, err
		}
		positions = pos
	}
	return types.Positions(positions).IntoProto(), nil
}
