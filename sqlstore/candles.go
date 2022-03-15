package sqlstore

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"code.vegaprotocol.io/data-node/candles"
	"github.com/shopspring/decimal"

	"code.vegaprotocol.io/data-node/entities"
	"github.com/georgysavva/scany/pgxscan"
)

const candlesViewIdentifier = "_candle_"

type Candles struct {
	*SQLStore
	candleIdToEventStream    map[string]*candleUpdatesStream
	subscriptionIdToCandleId map[uint64]string
	nextSubscriptionId       uint64
	sourceTableName          string
	subscriptionMutex        sync.Mutex
	config                   candles.Config
	ctx                      context.Context
}

func NewCandles(ctx context.Context, sqlStore *SQLStore, sourceTableName string, config candles.Config) (*Candles, error) {
	cs := &Candles{
		SQLStore:                 sqlStore,
		candleIdToEventStream:    map[string]*candleUpdatesStream{},
		subscriptionIdToCandleId: map[uint64]string{},
		sourceTableName:          sourceTableName,
		ctx:                      ctx,
		config:                   config,
	}

	for _, interval := range strings.Split(config.DefaultCandleIntervals, ",") {
		if interval != "" {
			viewAlreadyExists, _, err := cs.viewExistsForInterval(ctx, interval)
			if err != nil {
				return nil, fmt.Errorf("creating candle store:%w", err)
			}

			if !viewAlreadyExists {
				err = cs.createViewForInterval(ctx, interval)
				if err != nil {
					return nil, fmt.Errorf("creating candles store:%w", err)
				}
			}
		}
	}

	return cs, nil
}

// GetCandleDataForTimeSpan gets the candles for a given interval, from and to are optional
func (cs *Candles) GetCandleDataForTimeSpan(ctx context.Context, candleId string, from *time.Time, to *time.Time,
	p entities.Pagination) ([]entities.Candle, error) {

	candle, err := cs.candleFromCandleId(candleId)
	if err != nil {
		return nil, fmt.Errorf("getting candle data for time span:%w", err)
	}

	exists, err := cs.candleExists(ctx, candle)
	if err != nil {
		return nil, fmt.Errorf("getting candles for time span:%w", err)
	}

	if !exists {
		return nil, fmt.Errorf("no candle exists for candle id:%s", candleId)
	}

	var candles []entities.Candle

	query := fmt.Sprintf("SELECT period_start, open, close, high, low, volume, last_update_in_period FROM %s WHERE market_id = $1",
		candle.view)

	groupAsBytes, err := hex.DecodeString(candle.group)
	if err != nil {
		return nil, fmt.Errorf("invalid group:%w", err)
	}

	args := []interface{}{groupAsBytes}

	if from != nil {
		query = fmt.Sprintf("%s AND period_start >= %s", query, nextBindVar(&args, from))
	}

	if to != nil {
		query = fmt.Sprintf("%s AND period_start < %s", query, nextBindVar(&args, to))
	}

	query, args = orderAndPaginateQuery(query, []string{"period_start"}, p, args...)

	err = pgxscan.Select(ctx, cs.pool, &candles, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying candles: %w", err)
	}

	return candles, nil
}

// GetExistingCandlesForGroup returns a map of existing intervals to candle ids for the given market
func (cs *Candles) GetExistingCandlesForGroup(ctx context.Context, group string) (map[string]string, error) {

	result, err := cs.getIntervalToView(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting existing candles:%w", err)
	}

	candles := map[string]string{}
	for interval, viewName := range result {
		candles[interval] = cs.getCandleId(viewName, group)
	}
	return result, nil
}

func (cs *Candles) GetCandleIdForIntervalAndGroup(ctx context.Context, interval string, group string) (bool, string, error) {
	interval, err := cs.normaliseInterval(ctx, interval)
	if err != nil {
		return false, "", fmt.Errorf("invalid interval: %w", err)
	}

	viewAlreadyExists, existingInterval, err := cs.viewExistsForInterval(ctx, interval)
	if err != nil {
		return false, "", fmt.Errorf("checking for existing view: %w", err)
	}

	if viewAlreadyExists {
		return true, cs.getCandleId(existingInterval, group), nil
	}

	return false, "", nil
}

func (cs *Candles) CreateCandleForIntervalAndGroup(ctx context.Context, interval string, group string) (string, error) {
	interval, err := cs.normaliseInterval(ctx, interval)
	if err != nil {
		return "", fmt.Errorf("invalid interval: %w", err)
	}

	viewAlreadyExists, existingInterval, err := cs.viewExistsForInterval(ctx, interval)
	if err != nil {
		return "", fmt.Errorf("checking for existing view: %w", err)
	}
	if viewAlreadyExists {
		return "",
			fmt.Errorf("an equivalent candle for interval %s already exists.  Existing equivalent candle interval:%s", interval, existingInterval)
	}

	candle := cs.candleFromIntervalAndGroup(interval, group)
	err = cs.createViewForInterval(ctx, candle.interval)
	if err != nil {
		return "", fmt.Errorf("creating view for interval:%w", err)
	}

	return candle.id, nil
}

func (cs *Candles) ValidateInterval(ctx context.Context, interval string) error {
	_, err := cs.normaliseInterval(ctx, interval)
	if err != nil {
		return fmt.Errorf("invalid interval:%w", err)
	}

	return nil
}

func (cs *Candles) getIntervalToView(ctx context.Context) (map[string]string, error) {

	query := fmt.Sprintf("SELECT view_name FROM timescaledb_information.continuous_aggregates where view_name like '%s%%'",
		cs.getViewPrePend())
	rows, err := cs.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("fetching existing views for interval: %w", err)
	}

	var viewNames []string
	for rows.Next() {
		var viewName string
		err := rows.Scan(&viewName)
		if err != nil {
			return nil, fmt.Errorf("fetching existing views for interval: %w", err)
		}
		viewNames = append(viewNames, viewName)
	}

	result := map[string]string{}
	for _, viewName := range viewNames {
		interval, err := cs.getIntervalFromViewName(viewName)
		if err != nil {
			return nil, fmt.Errorf("fetching existing views for interval: %w", err)
		}

		result[interval] = viewName
	}
	return result, nil
}

func (cs *Candles) createViewForInterval(ctx context.Context, interval string) error {
	view := cs.getViewNameForInterval(interval)

	// timescaledb_experimental.time_bucket_ng for timescale 2.6.0
	query := fmt.Sprintf(`CREATE MATERIALIZED VIEW %s
		WITH (timescaledb.continuous) AS
		SELECT market_id,
			time_bucket('%s', synthetic_time) AS period_start, first(price, synthetic_time) AS open, last(price, synthetic_time) AS close,
			max(price) AS high, min(price) AS low, sum(size) AS volume, last(synthetic_time, synthetic_time) AS last_update_in_period
		FROM %s
		GROUP BY market_id, period_start`, view, interval, cs.sourceTableName)

	_, err := cs.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("creating view %s: %w", view, err)
	}

	return nil
}

func (cs *Candles) candleExists(ctx context.Context, candle candle) (bool, error) {

	exists, _, err := cs.viewExistsForInterval(ctx, candle.interval)
	if err != nil {
		return false, fmt.Errorf("candle exists:%w", err)
	}

	return exists, nil
}

func (cs *Candles) viewExistsForInterval(ctx context.Context, interval string) (bool, string, error) {
	intervalToView, err := cs.getIntervalToView(ctx)
	if err != nil {
		return false, "", fmt.Errorf("checking if view exist for interval:%w", err)
	}

	if _, ok := intervalToView[interval]; ok {
		return true, interval, nil
	}

	// Also check for existing Intervals that are specified differently but amount to the same thing  (i.e 7 days = 1 week)
	existingIntervals := map[int64]string{}
	for existingInterval := range intervalToView {
		seconds, err := cs.getIntervalSeconds(ctx, existingInterval)
		if err != nil {
			return false, "", fmt.Errorf("checking if view exists for interval:%w", err)
		}
		existingIntervals[seconds] = existingInterval
	}

	seconds, err := cs.getIntervalSeconds(ctx, interval)
	if existingInterval, ok := existingIntervals[seconds]; ok {
		return true, existingInterval, nil
	}

	return false, "", nil
}

// subscribe to a channel of new or updated candles. The subscriber id will be returned as an uint64 value
// and must be retained for future reference and to unsubscribe.
func (cs *Candles) subscribe(ctx context.Context, candleId string) (uint64, <-chan entities.Candle, error) {
	cs.subscriptionMutex.Lock()
	defer cs.subscriptionMutex.Unlock()

	candle, err := cs.candleFromCandleId(candleId)
	if err != nil {
		return 0, nil, fmt.Errorf("getting candles for time span:%w", err)
	}

	exists, err := cs.candleExists(ctx, candle)
	if err != nil {
		return 0, nil, fmt.Errorf("getting candles for time span:%w", err)
	}

	if !exists {
		return 0, nil, fmt.Errorf("no candle exists for candle id:%s", candleId)
	}

	if _, ok := cs.candleIdToEventStream[candleId]; !ok {
		evtStream, err := newCandleUpdatesStream(cs.ctx, cs.log, candleId, cs, cs.config)
		if err != nil {
			return 0, nil, fmt.Errorf("subsribing to candle updates:%w", err)
		}

		cs.candleIdToEventStream[candleId] = evtStream
	}

	evtStream := cs.candleIdToEventStream[candleId]
	cs.nextSubscriptionId++
	subscriptionId := cs.nextSubscriptionId

	out := evtStream.subscribe(subscriptionId)

	return subscriptionId, out, nil
}

func (cs *Candles) unsubscribe(subscriptionId uint64) error {
	cs.subscriptionMutex.Lock()
	defer cs.subscriptionMutex.Unlock()

	if candleId, ok := cs.subscriptionIdToCandleId[subscriptionId]; ok {
		evtStream := cs.candleIdToEventStream[candleId]
		evtStream.unsubscribe(subscriptionId)
		return nil
	} else {
		return fmt.Errorf("no subscription with id %d found", subscriptionId)
	}
}

func (cs *Candles) normaliseInterval(ctx context.Context, interval string) (string, error) {

	var normalizedInterval string

	_, err := cs.pool.Exec(ctx, "SET intervalstyle = 'postgres_verbose' ")
	if err != nil {
		return "", fmt.Errorf("normalising interval, failed to set interval style:%w", err)
	}

	query := fmt.Sprintf("select cast( INTERVAL '%s' as text)", interval)
	row := cs.pool.QueryRow(ctx, query)

	err = row.Scan(&normalizedInterval)
	if err != nil {
		return "", fmt.Errorf("normalising interval:%s :%w", interval, err)
	}

	normalizedInterval = strings.ReplaceAll(normalizedInterval, "@ ", "")

	return normalizedInterval, nil
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

type candle struct {
	id       string
	view     string
	interval string
	group    string
}

func (cs *Candles) candleFromCandleId(id string) (candle, error) {
	group, view, err := getGroupAndViewFromCandleId(id)
	if err != nil {
		return candle{}, fmt.Errorf("parsing candle id, failed to get group and view:%w", err)
	}

	split := strings.Split(view, cs.getViewPrePend())
	if len(split) != 2 {
		return candle{}, fmt.Errorf("parsing candle id, view name has unexpected format:%s", id)
	}

	interval, err := cs.getIntervalFromViewName(view)
	if err != nil {
		return candle{}, fmt.Errorf("parsing candle id, failed to get interval from view name:%w", err)
	}

	return candle{
		id:       id,
		view:     view,
		interval: interval,
		group:    group,
	}, nil

}

func (cs *Candles) candleFromIntervalAndGroup(interval string, group string) candle {
	view := cs.getViewNameForInterval(interval)
	id := view + "_" + group

	return candle{
		id:       id,
		view:     view,
		interval: interval,
		group:    group,
	}
}

func (cs *Candles) getViewPrePend() string {
	return cs.sourceTableName + candlesViewIdentifier
}

func (cs *Candles) getIntervalFromViewName(viewName string) (string, error) {
	split := strings.Split(viewName, cs.getViewPrePend())
	if len(split) != 2 {
		return "", fmt.Errorf("view name has unexpected format:%s", viewName)
	}
	return strings.ReplaceAll(split[1], "_", " "), nil
}

func (cs *Candles) getViewNameForInterval(interval string) string {
	return cs.getViewPrePend() + strings.ReplaceAll(interval, " ", "_")
}

func (cs *Candles) getCandleId(interval string, group string) string {
	view := cs.getViewNameForInterval(interval)
	return view + "_" + group
}

func getGroupAndViewFromCandleId(candleId string) (string, string, error) {
	idx := strings.LastIndex(candleId, "_")

	if idx == -1 {
		return "", "", fmt.Errorf("invalid candle id:%s", candleId)
	}
	return candleId[idx+1:], candleId[:idx], nil
}
