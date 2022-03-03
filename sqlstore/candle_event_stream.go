package sqlstore

import (
	"code.vegaprotocol.io/data-node/logging"
	"context"
	"fmt"
	"time"

	"code.vegaprotocol.io/data-node/entities"
)

type periodBoundarySource interface {
	getCandlePeriodBoundaries(ctx context.Context, interval string, from time.Time, numBoundaries int) ([]time.Time, error)
}

type candleEventStream struct {
	log         *logging.Logger
	interval    string
	marketID    []byte
	candleStore periodBoundarySource

	subscribers map[uint64]chan entities.Candle
	baseCandle  entities.Candle

	periodBoundariesFetchSize int
	periodBoundaries          []time.Time

	candleEventsChanSize int
}

func newCandleEventStream(log *logging.Logger, baseCandle entities.Candle, interval string, marketId []byte, periodBoundarySource periodBoundarySource,
	periodBoundariesFetchSize int, candleEventChanSize int) *candleEventStream {

	return &candleEventStream{
		log:                       log,
		baseCandle:                baseCandle,
		interval:                  interval,
		marketID:                  marketId,
		candleStore:               periodBoundarySource,
		periodBoundariesFetchSize: periodBoundariesFetchSize,
		subscribers:               map[uint64]chan entities.Candle{},
		candleEventsChanSize:      candleEventChanSize,
	}

}

func (s *candleEventStream) subscribe(subscriberId uint64) <-chan entities.Candle {
	out := make(chan entities.Candle, s.candleEventsChanSize)
	s.subscribers[subscriberId] = out
	return out
}

func (s *candleEventStream) unsubscribe(subscriberId uint64) {
	if out, ok := s.subscribers[subscriberId]; ok {
		delete(s.subscribers, subscriberId)
		close(out)
	}
}

func (s *candleEventStream) sendCandles(ctx context.Context, trades []*entities.Trade) error {

	if len(trades) == 0 {
		return nil
	}

	candles, err := s.applyTrades(ctx, s.baseCandle, trades)
	if err != nil {
		return fmt.Errorf("failed to send candles:%w", err)
	}

	if len(candles) > 0 {
		s.baseCandle = candles[len(candles)-1]
	}

	var slowConsumers []uint64
	for _, candle := range candles {
		for subscriberId, outCh := range s.subscribers {
			if len(outCh) < cap(outCh) {
				outCh <- candle
			} else {
				slowConsumers = append(slowConsumers, subscriberId)
			}
		}
	}

	for _, slowConsumer := range slowConsumers {
		s.log.Warningf("slow consumer detected, unsubscribing subscription id %d", slowConsumer)
		s.unsubscribe(slowConsumer)
	}

	return nil
}

func (s *candleEventStream) applyTrades(ctx context.Context, baseCandle entities.Candle,
	trades []*entities.Trade) ([]entities.Candle, error) {

	var candles []entities.Candle
	for _, trade := range trades {

		candlePeriodForTrade, err := s.getCandlePeriodForTrade(ctx, trade)
		if err != nil {
			return nil, fmt.Errorf("failed to apply trades:%w", err)
		}

		if baseCandle.Volume == 0 || candlePeriodForTrade.After(baseCandle.PeriodStart) {
			baseCandle = entities.Candle{
				PeriodStart:        s.periodBoundaries[0],
				LastUpdateInPeriod: trade.SyntheticTime,
				Open:               trade.Price,
				Close:              trade.Price,
				High:               trade.Price,
				Low:                trade.Price,
				Volume:             trade.Size,
			}
		} else {
			baseCandle.LastUpdateInPeriod = trade.SyntheticTime
			baseCandle.Volume += trade.Size
			if trade.Price.GreaterThan(baseCandle.High) {
				baseCandle.High = trade.Price
			}

			if trade.Price.LessThan(baseCandle.Low) {
				baseCandle.Low = trade.Price
			}

			baseCandle.Close = trade.Price
		}

		if len(candles) == 0 || baseCandle.PeriodStart.After(candles[len(candles)-1].PeriodStart) {
			candles = append(candles, baseCandle)
		} else {
			candles[len(candles)-1] = baseCandle
		}
	}

	return candles, nil
}

func (s *candleEventStream) getCandlePeriodForTrade(ctx context.Context, trade *entities.Trade) (time.Time, error) {

	if s.tradeTimeIsAfterLastPeriodBoundary(trade) {
		err := s.updatePeriodBoundaries(ctx, trade)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to get candle period for trade:%w", err)
		}
	}

	candlePeriod := s.periodBoundaries[0]
	for !trade.SyntheticTime.Before(s.periodBoundaries[1]) {
		s.periodBoundaries = s.periodBoundaries[1:]
		candlePeriod = s.periodBoundaries[0]
	}

	return candlePeriod, nil
}

func (s *candleEventStream) tradeTimeIsAfterLastPeriodBoundary(trade *entities.Trade) bool {
	return len(s.periodBoundaries) == 0 ||
		trade.SyntheticTime.After(s.periodBoundaries[len(s.periodBoundaries)-1])
}

func (s *candleEventStream) updatePeriodBoundaries(ctx context.Context, trade *entities.Trade) error {

	var err error
	s.periodBoundaries, err = s.candleStore.getCandlePeriodBoundaries(ctx, s.interval, trade.SyntheticTime,
		s.periodBoundariesFetchSize)
	if err != nil {
		return fmt.Errorf("unable to update period boundaries: %w", err)
	}

	if s.tradeTimeIsAfterLastPeriodBoundary(trade) {
		return fmt.Errorf("trade time exceeds the last period boundary after period boundaries have been updated")
	}

	return nil
}
