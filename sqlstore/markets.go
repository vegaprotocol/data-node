package sqlstore

import (
	"context"
	"fmt"
	"sync"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/logging"
	"github.com/georgysavva/scany/pgxscan"
)

type Markets struct {
	*ConnectionSource
	cache     map[string]entities.Market
	cacheLock sync.Mutex
}

const (
	sqlMarketsColumns = `id, vega_time, instrument_id, tradable_instrument, decimal_places,
		fees, opening_auction, price_monitoring_settings, liquidity_monitoring_parameters,
		trading_mode, state, market_timestamps, position_decimal_places`
)

func NewMarkets(connectionSource *ConnectionSource) *Markets {
	return &Markets{
		ConnectionSource: connectionSource,
		cache:            make(map[string]entities.Market),
	}
}

func (m *Markets) Upsert(ctx context.Context, market *entities.Market) error {
	m.cacheLock.Lock()
	defer m.cacheLock.Unlock()
	m.log.Info("adding market", logging.String("id", market.ID.String()))
	if market.Fees.Factors == nil {
		m.log.Info("insert ON NO NIL FACTORS", logging.String("id", market.ID.String()))
	} else {
		m.log.Info("insert non nil factors", logging.String("id", market.ID.String()))
	}

	query := fmt.Sprintf(`insert into markets(%s) 
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
on conflict (id, vega_time) do update
set 
	instrument_id=EXCLUDED.instrument_id,
	tradable_instrument=EXCLUDED.tradable_instrument,
	decimal_places=EXCLUDED.decimal_places,
	fees=EXCLUDED.fees,
	opening_auction=EXCLUDED.opening_auction,
	price_monitoring_settings=EXCLUDED.price_monitoring_settings,
	liquidity_monitoring_parameters=EXCLUDED.liquidity_monitoring_parameters,
	trading_mode=EXCLUDED.trading_mode,
	state=EXCLUDED.state,
	market_timestamps=EXCLUDED.market_timestamps,
	position_decimal_places=EXCLUDED.position_decimal_places;`, sqlMarketsColumns)

	if _, err := m.Connection.Exec(ctx, query, market.ID, market.VegaTime, market.InstrumentID, market.TradableInstrument, market.DecimalPlaces,
		market.Fees, market.OpeningAuction, market.PriceMonitoringSettings, market.LiquidityMonitoringParameters,
		market.TradingMode, market.State, market.MarketTimestamps, market.PositionDecimalPlaces); err != nil {
		err = fmt.Errorf("could not insert market into database: %w", err)
		return err
	}

	m.cache[market.ID.String()] = *market
	return nil
}

func (m *Markets) GetByID(ctx context.Context, marketID string) (entities.Market, error) {
	m.cacheLock.Lock()
	defer m.cacheLock.Unlock()
	m.log.Info("calling SQLstore GetById", logging.String("id", marketID))
	var market entities.Market

	if market, ok := m.cache[marketID]; ok {
		m.log.Info("found in cache", logging.String("id", marketID))
		if market.Fees.Factors == nil {
			m.log.Info("cache ON NO NIL FACTORS", logging.String("id", market.ID.String()))
		} else {
			m.log.Info("cache non nil factors", logging.String("id", market.ID.String()))
		}
		return market, nil
	}

	query := fmt.Sprintf(`select distinct on (id) %s 
from markets 
where id = $1
order by id, vega_time desc
`, sqlMarketsColumns)
	err := pgxscan.Get(ctx, m.Connection, &market, query, entities.NewMarketID(marketID))
	if err != nil {
		m.cache[marketID] = market
	}
	m.log.Info("found in DB", logging.String("id", marketID))
	return market, err
}

func (m *Markets) GetAll(ctx context.Context, pagination entities.Pagination) ([]entities.Market, error) {
	var markets []entities.Market
	query := fmt.Sprintf(`select distinct on (id) %s
from markets
order by id, vega_time desc
`, sqlMarketsColumns)

	query, _ = orderAndPaginateQuery(query, nil, pagination)

	err := pgxscan.Select(ctx, m.Connection, &markets, query)

	return markets, err
}
