package sqlstore

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"code.vegaprotocol.io/data-node/candles"

	"github.com/shopspring/decimal"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/logging"
	"github.com/georgysavva/scany/pgxscan"
)

type Candles struct {
	*SQLStore
	candleIdToEventStream     map[string]*candleEventStream
	marketIdToEventStreams    map[string][]*candleEventStream
	subscriptionIdToCandleId  map[uint64]string
	nextSubscriberId          uint64
	candleEventChanSize       int
	periodBoundariesFetchSize int
}

func NewCandles(sqlStore *SQLStore, config candles.Config) *Candles {
	c := &Candles{
		SQLStore:                  sqlStore,
		subscriptionIdToCandleId:  map[uint64]string{},
		candleIdToEventStream:     map[string]*candleEventStream{},
		marketIdToEventStreams:    map[string][]*candleEventStream{},
		candleEventChanSize:       config.CandleEventStreamBufferSize,
		periodBoundariesFetchSize: config.PeriodFetchSize,
	}

	return c
}

// GetCandlesForInterval gets the candles for a given interval, from and to are optional
func (cs *Candles) GetCandlesForInterval(ctx context.Context, interval string, from *time.Time, to *time.Time, marketId []byte,
	p entities.Pagination) ([]entities.Candle, error,
) {
	err := cs.ValidateInterval(ctx, interval)
	if err != nil {
		return nil, fmt.Errorf("invalid interval: %w", err)
	}

	var candles []entities.Candle

	query := `SELECT time_bucket($1, synthetic_time) AS period_start, first(price, synthetic_time) AS open, last(price, synthetic_time) AS close,
      max(price) AS high, min(price) AS low, sum(size) AS volume, last(synthetic_time, synthetic_time) AS last_update_in_period
	  FROM trades WHERE market_id = $2`

	args := []interface{}{fmt.Sprintf("'%s'", interval), marketId}

	if from != nil {
		query = fmt.Sprintf("%s AND synthetic_time >= %s", query, nextBindVar(&args, from))
	}

	if to != nil {
		query = fmt.Sprintf("%s AND synthetic_time < %s", query, nextBindVar(&args, to))
	}

	query = query + " GROUP BY period_start"
	query, args = orderAndPaginateQuery(query, []string{"period_start"}, p, args...)

	err = pgxscan.Select(ctx, cs.pool, &candles, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying candles: %w", err)
	}

	return candles, nil
}

// Subscribe to a channel of new or updated candles. The subscriber id will be returned as an uint64 value
// and must be retained for future reference and to unsubscribe.
func (cs *Candles) subscribe(ctx context.Context, marketIdBytes []byte, interval string) (uint64, <-chan entities.Candle, error) {
	err := cs.ValidateInterval(ctx, interval)
	if err != nil {
		return 0, nil, fmt.Errorf("invalid interval: %w", err)
	}

	cs.nextSubscriberId++
	subscriberId := cs.nextSubscriberId

	marketId := hex.EncodeToString(marketIdBytes)
	candleId := marketId + "-" + interval
	cs.subscriptionIdToCandleId[subscriberId] = candleId

	if _, ok := cs.candleIdToEventStream[candleId]; !ok {
		candles, err := cs.GetCandlesForInterval(ctx, interval, nil, nil, marketIdBytes, entities.Pagination{
			Skip:       0,
			Limit:      1,
			Descending: true,
		})
		if err != nil {
			return 0, nil, fmt.Errorf("failed to get last candle:%w", err)
		}
		var lastCandle entities.Candle
		if len(candles) > 0 {
			lastCandle = candles[len(candles)-1]
		} else {
			lastCandle = entities.Candle{}
		}

		evtStream := newCandleEventStream(cs.log, lastCandle, interval, marketIdBytes, cs,
			cs.periodBoundariesFetchSize, cs.candleEventChanSize)
		cs.candleIdToEventStream[candleId] = evtStream
		cs.marketIdToEventStreams[marketId] = append(cs.marketIdToEventStreams[marketId], evtStream)
	}

	evtStream := cs.candleIdToEventStream[candleId]

	out := evtStream.subscribe(subscriberId)
	cs.log.Info("Candle event stream subscriber added",
		logging.Uint64("subscriber-id", subscriberId), logging.String("candle-id", candleId))

	return subscriberId, out, nil
}

func (cs *Candles) unsubscribe(subscriberID uint64) error {
	if candleId, ok := cs.subscriptionIdToCandleId[subscriberID]; ok {
		evtStream := cs.candleIdToEventStream[candleId]
		evtStream.unsubscribe(subscriberID)
		cs.log.Info("Candle event stream subscriber removed",
			logging.Uint64("subscriber-id", subscriberID))
		return nil
	} else {
		return fmt.Errorf("no subscriber with id %d found", subscriberID)
	}
}

// getCandlePeriodBoundaries returns a list of period boundaries starting at the first period before `from`
func (cs *Candles) getCandlePeriodBoundaries(ctx context.Context, interval string, from time.Time, numBoundaries int) ([]time.Time, error) {
	intervalSeconds, err := cs.getIntervalSeconds(ctx, interval)
	if err != nil {
		return nil, fmt.Errorf("failed to get candle periods:%w", err)
	}

	to := from.Add(time.Duration(intervalSeconds*(int64(numBoundaries)-1)) * time.Second)

	var periodBoundaries []time.Time

	query := `SELECT time_bucket_gapfill($1, synthetic_time, $2 , $3) AS period_start
		  FROM trades
		  GROUP BY period_start
		  ORDER BY period_start ASC`

	args := []interface{}{fmt.Sprintf("'%s'", interval), from, to}

	rows, err := cs.pool.Query(ctx, query, args...)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to get candle period boundaries: %w", err)
	}
	for rows.Next() {
		t := time.Time{}
		err = rows.Scan(&t)
		if err != nil {
			return nil, fmt.Errorf("failed to get candle period boundaries: %w", err)
		}

		periodBoundaries = append(periodBoundaries, t)

	}

	return periodBoundaries, nil
}

func (cs *Candles) ValidateInterval(ctx context.Context, interval string) error {
	_, err := cs.getIntervalSeconds(ctx, interval)
	if err != nil {
		return fmt.Errorf("invalid interval:%s", err)
	}

	// The postgres time_bucket_ng api that supports year and months is marked as experimental, so intervals of this type
	// are not currently support.
	if strings.Contains(interval, "year") {
		return fmt.Errorf("interval does not support years")
	}

	if strings.Contains(interval, "mon") {
		return fmt.Errorf("interval does not support months")
	}

	return nil
}

func (cs *Candles) getIntervalSeconds(ctx context.Context, interval string) (int64, error) {
	var seconds decimal.Decimal

	query := fmt.Sprintf("SELECT EXTRACT(epoch FROM INTERVAL '%s')", interval)
	row := cs.pool.QueryRow(ctx, query)

	err := row.Scan(&seconds)
	if err != nil {
		return 0, err
	}

	return seconds.IntPart(), nil
}

func (cs *Candles) sendCandles(ctx context.Context, trades []*entities.Trade) error {
	marketToTrades := map[string][]*entities.Trade{}

	for _, trade := range trades {
		marketId := hex.EncodeToString(trade.MarketID)
		marketToTrades[marketId] = append(marketToTrades[marketId], trade)
	}

	for marketId, marketTrades := range marketToTrades {
		for _, evtStream := range cs.marketIdToEventStreams[marketId] {
			err := evtStream.sendCandles(ctx, marketTrades)
			if err != nil {
				return fmt.Errorf("failed to update event stream for market id:%s :%w", marketId, err)
			}
		}
	}

	return nil
}
