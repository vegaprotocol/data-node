package sqlstore

import (
	"code.vegaprotocol.io/data-node/candles"
	"context"
	"fmt"
	"time"

	"code.vegaprotocol.io/data-node/logging"

	"code.vegaprotocol.io/data-node/entities"
)

type candleSource interface {
	GetCandleDataForTimeSpan(ctx context.Context, candleId string, from *time.Time, to *time.Time,
		p entities.Pagination) ([]entities.Candle, error)
}

type candleUpdatesStream struct {
	log             *logging.Logger
	candleSource    candleSource
	candleId        string
	subscribeChan   chan subscribeRequest
	unsubscribeChan chan uint64
	config          candles.Config
}

func newCandleUpdatesStream(ctx context.Context, log *logging.Logger, candleId string, candleSource candleSource,
	config candles.Config) (*candleUpdatesStream, error) {

	ces := &candleUpdatesStream{
		log:             log,
		candleSource:    candleSource,
		candleId:        candleId,
		config:          config,
		subscribeChan:   make(chan subscribeRequest),
		unsubscribeChan: make(chan uint64),
	}

	go ces.run(ctx)

	return ces, nil
}

func (s *candleUpdatesStream) run(ctx context.Context) {
	subscriptions := map[uint64]chan entities.Candle{}
	defer closeAllSubscriptions(subscriptions)

	ticker := time.NewTicker(s.config.CandleUpdatesStreamInterval.Duration)
	var lastCandle *entities.Candle

	for {
		var err error
		select {
		case <-ctx.Done():
			ctx.Err()
			return
		case <-ticker.C:
			if len(subscriptions) > 0 {
				if lastCandle == nil {
					lastCandle, err = s.getLastCandle(ctx)
					if err != nil {
						s.log.Errorf("error whilst getting last candle, closing stream for candle id %s: %w", s.candleId, err)
						return
					}
				}

				if lastCandle != nil {
					candles, err := s.getCandleUpdatesSinceLastCandle(ctx, lastCandle)
					if err != nil {
						s.log.Errorf("failed to get candles, closing stream for candle id %s: %w", s.candleId, err)
						return
					}

					if len(candles) > 0 {
						lastCandle = &candles[len(candles)-1]
					}

					slowConsumers, err := s.sendCandles(candles, subscriptions)
					if err != nil {
						s.log.Errorf("failed to send candles, closing stream for candle id %s: %w", s.candleId, err)
						return
					}

					for _, slowConsumerId := range slowConsumers {
						s.log.Warningf("slow consumer detected, unsubscribing subscription id %d", slowConsumerId)
						removeSubscription(subscriptions, slowConsumerId)
					}
				}
			}
		case subscription := <-s.subscribeChan:
			if len(subscriptions) == 0 {
				lastCandle, err = s.getLastCandle(ctx)
				if err != nil {
					s.log.Errorf("error whilst getting last candle, closing stream for candle id %s: %w", s.candleId, err)
					return
				}
			}

			subscriptions[subscription.id] = subscription.out
			s.log.Info("Candle updates subscription added",
				logging.Uint64("subscription-id", subscription.id), logging.String("candle-id", s.candleId))
		case id := <-s.unsubscribeChan:
			removeSubscription(subscriptions, id)
			s.log.Info("Candle updates subscription removed",
				logging.Uint64("subscription-id", id))
		}
	}
}

func removeSubscription(subscriptions map[uint64]chan entities.Candle, subscriptionId uint64) {
	if _, ok := subscriptions[subscriptionId]; ok {
		close(subscriptions[subscriptionId])
		delete(subscriptions, subscriptionId)
	}
}

func closeAllSubscriptions(subscribers map[uint64]chan entities.Candle) {
	for _, subscriber := range subscribers {
		close(subscriber)
	}
}

type subscribeRequest struct {
	id  uint64
	out chan entities.Candle
}

func (s *candleUpdatesStream) subscribe(subscriberId uint64) <-chan entities.Candle {

	out := make(chan entities.Candle, s.config.CandleUpdatesStreamBufferSize)
	s.subscribeChan <- subscribeRequest{
		id:  subscriberId,
		out: out,
	}

	return out
}

func (s *candleUpdatesStream) unsubscribe(subscriberId uint64) {
	s.unsubscribeChan <- subscriberId
}

func (s *candleUpdatesStream) getCandleUpdatesSinceLastCandle(ctx context.Context, lastCandle *entities.Candle) ([]entities.Candle, error) {
	ctx, _ = context.WithTimeout(ctx, s.config.CandlesFetchTimeout.Duration)
	if lastCandle != nil {
		start := lastCandle.PeriodStart
		candles, err := s.candleSource.GetCandleDataForTimeSpan(ctx, s.candleId, &start, nil, entities.Pagination{})
		if err != nil {
			return nil, fmt.Errorf("getting candles:%w", err)
		}

		var updates []entities.Candle
		for _, candle := range candles {
			if candle.LastUpdateInPeriod.After(lastCandle.LastUpdateInPeriod) {
				updates = append(updates, candle)
			}
		}

		return updates, nil
	}

	return nil, nil
}

func (s *candleUpdatesStream) sendCandles(candles []entities.Candle, subscribers map[uint64]chan entities.Candle) ([]uint64, error) {

	var slowConsumers []uint64
	for _, candle := range candles {
		for subscriberId, outCh := range subscribers {
			if len(outCh) < cap(outCh) {
				outCh <- candle
			} else {
				slowConsumers = append(slowConsumers, subscriberId)
			}
		}
	}

	return slowConsumers, nil
}

func (s *candleUpdatesStream) getLastCandle(ctx context.Context) (*entities.Candle, error) {
	ctx, _ = context.WithTimeout(ctx, s.config.CandlesFetchTimeout.Duration)
	pagination := entities.Pagination{
		Skip:       0,
		Limit:      1,
		Descending: true,
	}
	candles, err := s.candleSource.GetCandleDataForTimeSpan(ctx, s.candleId, nil, nil, pagination)
	if err != nil {
		return nil, fmt.Errorf("getting last candle:%w", err)
	}

	if len(candles) == 0 {
		return nil, nil
	}

	if len(candles) == 1 {
		return &candles[0], nil
	}

	return nil, nil
}
