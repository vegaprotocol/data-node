package sqlstore

import (
	"context"
	"fmt"

	"code.vegaprotocol.io/data-node/entities"
)

type MarketData struct {
	*SqlStore
}

func NewMarketData(sqlStore *SqlStore) *MarketData {
	return &MarketData{
		SqlStore: sqlStore,
	}
}

func (md *MarketData) Add(ctx context.Context, data *entities.MarketData) error {
	tx, err := md.pool.Begin(ctx)
	if err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, `insert into market_data (
	market, market_timestamp, vega_time, mark_price, 
	best_bid_price, best_bid_volume, best_offer_price, best_offer_volume,
	best_static_bid_price, best_static_bid_volume, best_static_offer_price, best_static_offer_volume,
	mid_price, static_mid_price, open_interest, auction_end, 
	auction_start, indicative_price, indicative_volume,	market_trading_mode, 
	auction_trigger, extension_trigger, target_stake, supplied_stake, 
	price_monitoring_bounds, market_value_proxy, liquidity_provider_fee_shares)
	values ($1, $2, $3, $4, 
			$5, $6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15, $16, 
			$17, $18, $19, $20,
			$21, $22, $23, $24,
			$25, $26, $27)`,
		data.Market, data.Timestamp, data.VegaTime, data.MarkPrice,
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

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("could not insert market data into the database, commit failed: %w", err)
	}

	return nil
}
