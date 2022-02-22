package sqlstore

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"code.vegaprotocol.io/data-node/entities"
	"github.com/georgysavva/scany/pgxscan"
)

type MarketData struct {
	*SQLStore
}

const (
	sqlColumns = `market, market_timestamp, vega_time, mark_price, 
		best_bid_price, best_bid_volume, best_offer_price, best_offer_volume,
		best_static_bid_price, best_static_bid_volume, best_static_offer_price, best_static_offer_volume,
		mid_price, static_mid_price, open_interest, auction_end, 
		auction_start, indicative_price, indicative_volume,	market_trading_mode, 
		auction_trigger, extension_trigger, target_stake, supplied_stake, 
		price_monitoring_bounds, market_value_proxy, liquidity_provider_fee_shares`
)

var ErrInvalidDateRange = errors.New("invalid date range, end date must be after start date")

func NewMarketData(sqlStore *SQLStore) *MarketData {
	return &MarketData{
		SQLStore: sqlStore,
	}
}

func (md *MarketData) Add(ctx context.Context, entries ...*entities.MarketData) error {
	tx, err := md.pool.Begin(ctx)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`insert into market_data(%s) 
	values ($1, $2, $3, $4, 
			$5, $6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15, $16, 
			$17, $18, $19, $20,
			$21, $22, $23, $24,
			$25, $26, $27)`, sqlColumns)
	for _, data := range entries {
		if _, err = tx.Exec(ctx, query,
			data.Market, data.MarketTimestamp, data.VegaTime, data.MarkPrice,
			data.BestBidPrice, data.BestBidVolume, data.BestOfferPrice, data.BestOfferVolume,
			data.BestStaticBidPrice, data.BestStaticBidVolume, data.BestStaticOfferPrice, data.BestStaticOfferVolume,
			data.MidPrice, data.StaticMidPrice, data.OpenInterest, data.AuctionEnd,
			data.AuctionStart, data.IndicativePrice, data.IndicativeVolume, data.MarketTradingMode,
			data.AuctionTrigger, data.ExtensionTrigger, data.TargetStake, data.SuppliedStake,
			data.PriceMonitoringBounds, data.MarketValueProxy, data.LiquidityProviderFeeShares,
		); err != nil {
			err = fmt.Errorf("could not insert into database: %w", err)
			if txErr := tx.Rollback(ctx); txErr != nil {
				return fmt.Errorf("rollback failed: %s, %w", txErr, err)
			}

			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("could not insert market data into the database, commit failed: %w", err)
	}

	return nil
}

func (md *MarketData) GetByID(ctx context.Context, marketID string) (entities.MarketData, error) {
	market, err := hex.DecodeString(marketID)

	if err != nil {
		return entities.MarketData{}, fmt.Errorf("bad ID (must be a hex string): %w", err)
	}

	var marketData entities.MarketData
	query := fmt.Sprintf("select %s from market_data_snapshot where market = $1", sqlColumns)

	err = pgxscan.Get(ctx, md.pool, &marketData, query, market)

	return marketData, err
}

func (md *MarketData) GetAll(ctx context.Context) (results []entities.MarketData, err error) {
	query := fmt.Sprintf(`select %s from market_data_snapshot`, sqlColumns)
	err = pgxscan.Select(ctx, md.pool, &results, query)
	return
}

func (md *MarketData) GetBetweenDatesByID(ctx context.Context, marketID string, start, end time.Time) ([]entities.MarketData, error) {
	if end.Before(start) {
		return nil, ErrInvalidDateRange
	}

	return md.getBetweenDatesByID(ctx, marketID, &start, &end)
}

func (md *MarketData) GetFromDateByID(ctx context.Context, marketID string, start time.Time) ([]entities.MarketData, error) {
	return md.getBetweenDatesByID(ctx, marketID, &start, nil)
}

func (md *MarketData) GetToDateByID(ctx context.Context, marketID string, end time.Time) ([]entities.MarketData, error) {
	return md.getBetweenDatesByID(ctx, marketID, nil, &end)
}

func (md *MarketData) getBetweenDatesByID(ctx context.Context, marketID string, start, end *time.Time) (results []entities.MarketData, err error) {
	var market []byte

	market, err = hex.DecodeString(marketID)
	if err != nil {
		return nil, err
	}

	selectStatement := fmt.Sprintf(`select %s from market_data`, sqlColumns)

	if start != nil && end != nil {
		query := fmt.Sprintf(`%s where market = $1 and market_timestamp between $2 and $3`, selectStatement)
		err = pgxscan.Select(ctx, md.pool, &results, query, market, *start, *end)
	} else if start != nil && end == nil {
		query := fmt.Sprintf(`%s where market = $1 and market_timestamp >= $2`, selectStatement)
		err = pgxscan.Select(ctx, md.pool, &results, query, market, *start)
	} else if start == nil && end != nil {
		query := fmt.Sprintf(`%s where market = $1 and market_timestamp <= $2`, selectStatement)
		err = pgxscan.Select(ctx, md.pool, &results, query, market, *end)
	}

	return results, err
}

func (md *MarketData) GetAllBetweenDates(ctx context.Context, start, end time.Time) ([]entities.MarketData, error) {
	if end.Before(start) {
		return nil, ErrInvalidDateRange
	}

	selectStatement := fmt.Sprintf("select %s from market_data", sqlColumns)
	query := fmt.Sprintf(`%s where market_timestamp between $1 and $2`, selectStatement)
	var results []entities.MarketData
	err := pgxscan.Select(ctx, md.pool, &results, query, start, end)

	return results, err
}
