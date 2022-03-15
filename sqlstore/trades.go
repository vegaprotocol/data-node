package sqlstore

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	"code.vegaprotocol.io/data-node/entities"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type Trades struct {
	*SQLStore
	candlesStore *Candles
	trades       []*entities.Trade

	mu sync.Mutex
}

func NewTrades(sqlStore *SQLStore, candlesStore *Candles) *Trades {
	t := &Trades{
		SQLStore:     sqlStore,
		candlesStore: candlesStore,
	}
	return t
}

func (ts *Trades) SubscribeToTradesCandle(ctx context.Context, candleId string) (uint64, <-chan entities.Candle, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	return ts.candlesStore.subscribe(ctx, candleId)
}

func (ts *Trades) UnsubscribeFromTradesCandle(subscriberID uint64) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	return ts.candlesStore.unsubscribe(subscriberID)
}

func (ts *Trades) OnTimeUpdateEvent(ctx context.Context) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	var rows [][]interface{}
	for _, t := range ts.trades {
		rows = append(rows, []interface{}{
			t.SyntheticTime,
			t.VegaTime,
			t.SeqNum,
			t.ID,
			t.MarketID,
			t.Price,
			t.Size,
			t.Buyer,
			t.Seller,
			t.Aggressor,
			t.BuyOrder,
			t.SellOrder,
			t.Type,
			t.BuyerMakerFee,
			t.BuyerInfrastructureFee,
			t.BuyerLiquidityFee,
			t.SellerMakerFee,
			t.SellerInfrastructureFee,
			t.SellerLiquidityFee,
			t.BuyerAuctionBatch,
			t.SellerAuctionBatch,
		})
	}

	if rows != nil {
		copyCount, err := ts.pool.CopyFrom(
			ctx,
			pgx.Identifier{"trades"},
			[]string{
				"synthetic_time", "vega_time", "seq_num", "id", "market_id", "price", "size", "buyer", "seller",
				"aggressor", "buy_order", "sell_order", "type", "buyer_maker_fee", "buyer_infrastructure_fee",
				"buyer_liquidity_fee", "seller_maker_fee", "seller_infrastructure_fee", "seller_liquidity_fee",
				"buyer_auction_batch", "seller_auction_batch",
			},
			pgx.CopyFromRows(rows),
		)
		if err != nil {
			return fmt.Errorf("failed to copy trades into database:%w", err)
		}

		if copyCount != int64(len(rows)) {
			return fmt.Errorf("copied %d rows into the database, expected to copy %d", copyCount, len(rows))
		}
	}

	ts.trades = nil

	return nil
}

func (ts *Trades) Add(t *entities.Trade) error {
	ts.trades = append(ts.trades, t)
	return nil
}

func (ts *Trades) GetByMarket(ctx context.Context, market string, p entities.Pagination) ([]entities.Trade, error) {
	marketId, err := hex.DecodeString(market)
	if err != nil {
		return nil, err
	}

	query := `SELECT * from trades WHERE market_id=$1`
	args := []interface{}{marketId}
	trades, err := ts.queryTrades(ctx, query, args, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to get trade by market:%w", err)
	}

	return trades, nil
}

func (ts *Trades) GetByParty(ctx context.Context, party string, market *string, pagination entities.Pagination) ([]entities.Trade, error) {
	partyId, err := hex.DecodeString(party)
	if err != nil {
		return nil, err
	}
	args := []interface{}{partyId}
	query := `SELECT * from trades WHERE buyer=$1 or seller=$1`

	return ts.queryTradesWithMarketFilter(ctx, query, args, market, pagination)
}

func (ts *Trades) GetByOrderID(ctx context.Context, order string, market *string, pagination entities.Pagination) ([]entities.Trade, error) {
	orderId, err := hex.DecodeString(order)
	if err != nil {
		return nil, err
	}
	args := []interface{}{orderId}
	query := `SELECT * from trades WHERE buy_order=$1 or sell_order=$1`
	return ts.queryTradesWithMarketFilter(ctx, query, args, market, pagination)
}

func (ts *Trades) queryTradesWithMarketFilter(ctx context.Context, query string, args []interface{}, market *string, p entities.Pagination) ([]entities.Trade, error) {
	if market != nil && *market != "" {

		marketId, err := hex.DecodeString(*market)
		if err != nil {
			return nil, err
		}
		position := nextBindVar(&args, marketId)
		query += ` AND market_id=` + position
	}

	trades, err := ts.queryTrades(ctx, query, args, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to query trades:%w", err)
	}

	return trades, nil
}

func (ts *Trades) queryTrades(ctx context.Context, query string, args []interface{}, p *entities.Pagination) ([]entities.Trade, error) {
	if p != nil {
		query, args = orderAndPaginateQuery(query, []string{"synthetic_time"}, *p, args...)
	}

	var trades []entities.Trade
	err := pgxscan.Select(ctx, ts.pool, &trades, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying trades: %w", err)
	}
	return trades, nil
}
