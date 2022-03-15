package sqlstore_test

import (
	"code.vegaprotocol.io/data-node/config/encoding"
	"context"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
	"time"

	"code.vegaprotocol.io/data-node/candles"
	types "code.vegaprotocol.io/protos/vega"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/sqlstore"
)

const (
	Midnight3rdMarch2022InUnixTime = 1646265600
	totalBlocks                    = 1000
	tradesPerBlock                 = 5
	startPrice                     = 1
	size                           = 10
	priceIncrement                 = 1
	blockIntervalSeconds           = 10
	blockIntervalDur               = time.Duration(blockIntervalSeconds) * time.Second
)

func Test_CreatingCandleWithIntervalEquivalentToExistingCandleIntervalFails(t *testing.T) {
	defer testStore.DeleteEverything()
	config := newTestCandleConfig(1)

	candleStore, _ := sqlstore.NewCandles(context.Background(), testStore, "trades", config)

	_, err := candleStore.CreateCandleForIntervalAndGroup(context.Background(), "7 days", testMarket)
	if err == nil {
		t.Fatalf("expected to fail as interval for 1 week exists")
	}

}

func Test_GetExistingCandles(t *testing.T) {
	defer testStore.DeleteEverything()
	config := newTestCandleConfig(1)

	candleStore, _ := sqlstore.NewCandles(context.Background(), testStore, "trades", config)
	candles, err := candleStore.GetExistingCandlesForGroup(context.Background(), testMarket)
	if err != nil {
		t.Fatalf("failed to get candles for market:%s", err)
	}

	intervals := strings.Split(config.DefaultCandleIntervals, ",")
	assert.Equal(t, len(intervals), len(candles))

	for _, interval := range intervals {
		_, ok := candles[interval]
		assert.True(t, ok)
	}

}

func Test_CandlesPagination(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore, err := sqlstore.NewCandles(context.Background(), testStore, "trades", newTestCandleConfig(1))
	if err != nil {
		t.Fatalf("failed to create candles store:%s", err)
	}

	tradeStore := sqlstore.NewTrades(testStore, candleStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	insertCandlesTestData(t, tradeStore, startTime, totalBlocks, tradesPerBlock, startPrice, priceIncrement, size, blockIntervalDur)

	_, candleId, _ := candleStore.GetCandleIdForIntervalAndGroup(context.Background(), "1 Minute", testMarket)
	candles, err := candleStore.GetCandleDataForTimeSpan(context.Background(), candleId, nil,
		nil, entities.Pagination{
			Skip:       0,
			Limit:      10,
			Descending: true,
		})
	if err != nil {
		t.Fatalf("failed to get candles with pagination:%s", err)
	}

	assert.Equal(t, 10, len(candles))
	lastCandle := candles[9]

	candles, err = candleStore.GetCandleDataForTimeSpan(context.Background(), candleId, nil,
		nil, entities.Pagination{
			Skip:       9,
			Limit:      5,
			Descending: true,
		})

	if err != nil {
		t.Fatalf("failed to get candles with pagination:%s", err)
	}

	assert.Equal(t, 5, len(candles))
	assert.Equal(t, lastCandle, candles[0])
}

func Test_CandlesGetForEmptyInterval(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore, err := sqlstore.NewCandles(context.Background(), testStore, "trades", newTestCandleConfig(1))
	if err != nil {
		t.Fatalf("failed to created candle store:%s", err)
	}
	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	insertTestTrade(t, tradeStore, 1, 10, block, 0)
	insertTestTrade(t, tradeStore, 2, 10, block, 3)

	nextTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0).Add(10 * time.Minute)
	block = addTestBlockForTime(t, bs, nextTime)
	insertTestTrade(t, tradeStore, 3, 20, block, 0)
	insertTestTrade(t, tradeStore, 4, 20, block, 5)

	_, candleId, err := candleStore.GetCandleIdForIntervalAndGroup(context.Background(), "1 Minute", testMarket)
	if err != nil {
		t.Fatalf("getting existing candle id:%s", err)
	}

	candles, err := candleStore.GetCandleDataForTimeSpan(context.Background(), candleId, &startTime,
		nil, entities.Pagination{})
	if err != nil {
		t.Fatalf("failed to get candles:%s", err)
	}

	assert.Equal(t, 2, len(candles))

	firstCandle := createCandle(startTime,
		startTime.Add(3*time.Microsecond), 1, 2, 2, 1, 20)
	assert.Equal(t, firstCandle, candles[0])

	secondCandle := createCandle(startTime.Add(10*time.Minute),
		startTime.Add(10*time.Minute).Add(5*time.Microsecond), 3, 4, 4, 3, 40)
	assert.Equal(t, secondCandle, candles[1])
}

func Test_CandlesGetLatest(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore, _ := sqlstore.NewCandles(context.Background(), testStore, "trades", newTestCandleConfig(1))
	tradeStore := sqlstore.NewTrades(testStore, candleStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	insertCandlesTestData(t, tradeStore, startTime, 90, 3, startPrice, priceIncrement, size,
		1*time.Second)

	_, candleId, _ := candleStore.GetCandleIdForIntervalAndGroup(context.Background(), "1 Minute", testMarket)
	candles, err := candleStore.GetCandleDataForTimeSpan(context.Background(), candleId, &startTime,
		nil, entities.Pagination{
			Skip:       0,
			Limit:      1,
			Descending: true,
		})
	if err != nil {
		t.Fatalf("failed to get candles:%s", err)
	}

	assert.Equal(t, 1, len(candles))

	lastCandle := createCandle(startTime.Add(60*time.Second),
		startTime.Add(89*time.Second).Add(2*time.Microsecond), 181, 270, 270, 181,
		900)
	assert.Equal(t, lastCandle, candles[0])
}

func Test_CandlesGetForDifferentIntervalAndTimeBounds(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore, _ := sqlstore.NewCandles(context.Background(), testStore, "trades", newTestCandleConfig(1))
	tradeStore := sqlstore.NewTrades(testStore, candleStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	insertCandlesTestData(t, tradeStore, startTime, totalBlocks, tradesPerBlock, startPrice, priceIncrement, size, blockIntervalDur)

	testInterval(t, startTime, nil, nil, candleStore, "1 Minute", 60)
	testInterval(t, startTime, nil, nil, candleStore, "5 Minutes", 300)
	testInterval(t, startTime, nil, nil, candleStore, "15 Minutes", 900)
	testInterval(t, startTime, nil, nil, candleStore, "1 hour", 3600)

	from := startTime.Add(5 * time.Minute)
	to := startTime.Add(35 * time.Minute)

	testInterval(t, startTime, &from, &to, candleStore, "1 Minute", 60)
	testInterval(t, startTime, &from, &to, candleStore, "5 Minutes", 300)

	testInterval(t, startTime, nil, &to, candleStore, "1 Minute", 60)
	testInterval(t, startTime, nil, &to, candleStore, "5 Minutes", 300)

	testInterval(t, startTime, &from, nil, candleStore, "1 Minute", 60)
	testInterval(t, startTime, &from, nil, candleStore, "5 Minutes", 300)
}

func newTestCandleConfig(bufferSize int) candles.Config {
	conf := candles.NewDefaultConfig()
	conf.CandleUpdatesStreamBufferSize = bufferSize
	conf.CandleUpdatesStreamInterval = encoding.Duration{Duration: time.Duration(1 * time.Microsecond)}
	conf.CandlesFetchTimeout = encoding.Duration{Duration: time.Duration(2 * time.Minute)}
	return conf
}

func testInterval(t *testing.T, tradeDataStartTime time.Time, fromTime *time.Time, toTime *time.Time, candleStore *sqlstore.Candles, interval string,
	intervalSeconds int) {
	intervalDur := time.Duration(intervalSeconds) * time.Second

	_, candleId, _ := candleStore.GetCandleIdForIntervalAndGroup(context.Background(), interval, testMarket)
	candles, err := candleStore.GetCandleDataForTimeSpan(context.Background(), candleId, fromTime,
		toTime, entities.Pagination{})
	if err != nil {
		t.Fatalf("failed to get candles:%s", err)
	}

	tradeDataTimeSpanSeconds := totalBlocks * blockIntervalSeconds
	tradeDataEndTime := tradeDataStartTime.Add(time.Duration(tradeDataTimeSpanSeconds) * time.Second)

	var candlesStartTime time.Time
	if fromTime != nil && fromTime.After(tradeDataStartTime) {
		candlesStartTime = *fromTime
	} else {
		candlesStartTime = tradeDataStartTime
	}

	var candlesEndTime time.Time
	if toTime != nil && toTime.Before(tradeDataEndTime) {
		candlesEndTime = *toTime
	} else {
		candlesEndTime = tradeDataEndTime
	}

	candleSpan := candlesEndTime.Sub(candlesStartTime)

	expectedNumCandles := int(candleSpan.Seconds() / float64(intervalSeconds))

	if toTime == nil {
		expectedNumCandles = expectedNumCandles + 1
	}

	assert.Equal(t, expectedNumCandles, len(candles))

	blocksPerInterval := int(intervalSeconds) / blockIntervalSeconds

	skippedTrades := int(candlesStartTime.Sub(tradeDataStartTime).Seconds()/blockIntervalSeconds) * tradesPerBlock

	for idx := 0; idx < expectedNumCandles-1; idx++ {
		periodStart := candlesStartTime.Add(time.Duration(idx) * intervalDur)
		tradesAtOpen := skippedTrades + idx*tradesPerBlock*blocksPerInterval
		tradesAtClose := skippedTrades + (idx+1)*tradesPerBlock*blocksPerInterval
		expectedVolume := tradesPerBlock * blocksPerInterval * size
		candle := candles[idx]
		expectedCandle := createCandle(periodStart,
			periodStart.Add(time.Duration(tradesPerBlock-1)*time.Microsecond).Add(intervalDur).Add(-1*blockIntervalDur),
			startPrice+tradesAtOpen, tradesAtClose, tradesAtClose, startPrice+tradesAtOpen,
			expectedVolume)
		assert.Equal(t, expectedCandle, candle)
	}
}

func createCandle(periodStart time.Time, lastUpdate time.Time, open int, close int, high int, low int, volume int) entities.Candle {
	return entities.Candle{
		PeriodStart:        periodStart,
		LastUpdateInPeriod: lastUpdate,
		Open:               decimal.NewFromInt(int64(open)),
		Close:              decimal.NewFromInt(int64(close)),
		High:               decimal.NewFromInt(int64(high)),
		Low:                decimal.NewFromInt(int64(low)),
		Volume:             uint64(volume),
	}
}

func insertCandlesTestData(t *testing.T, tradeStore *sqlstore.Trades, startTime time.Time, numBlocks int,
	tradePerBlock int, startPrice int, priceIncrement int, size int, blockIntervalDur time.Duration) {
	bs := sqlstore.NewBlocks(testStore)

	var blocks []entities.Block
	for i := 0; i < numBlocks; i++ {
		blocks = append(blocks, addTestBlockForTime(t, bs, startTime.Add(time.Duration(i)*blockIntervalDur)))
	}

	for _, block := range blocks {
		for seqNum := 0; seqNum < tradePerBlock; seqNum++ {
			trade := createTestTrade(t, startPrice, size, block, seqNum)
			err := tradeStore.Add(trade)
			if err != nil {
				t.Fatalf("failed to add trade to store:%s", err)
			}
			startPrice = startPrice + priceIncrement
		}
	}
	tradeStore.OnTimeUpdateEvent(context.Background())
}

func insertTestTrade(t *testing.T, tradeStore *sqlstore.Trades, price int, size int, block entities.Block, seqNum int) *entities.Trade {
	trade := createTestTrade(t, price, size, block, seqNum)
	return insertTrade(t, tradeStore, trade)
}

func insertTrade(t *testing.T, tradeStore *sqlstore.Trades, trade *entities.Trade) *entities.Trade {
	err := tradeStore.Add(trade)
	tradeStore.OnTimeUpdateEvent(context.Background())
	if err != nil {
		t.Fatalf("failed to add trade to store:%s", err)
	}

	return trade
}

func createTestTrade(t *testing.T, price int, size int, block entities.Block, seqNum int) *entities.Trade {
	proto := &types.Trade{
		Type:      types.Trade_TYPE_DEFAULT,
		Id:        hex.EncodeToString(generateID()),
		Price:     strconv.Itoa(price),
		Size:      uint64(size),
		MarketId:  testMarket,
		Buyer:     hex.EncodeToString(generateID()),
		Seller:    hex.EncodeToString(generateID()),
		Aggressor: types.Side_SIDE_SELL,
		BuyOrder:  hex.EncodeToString(generateID()),
		SellOrder: hex.EncodeToString(generateID()),
	}

	trade, err := entities.TradeFromProto(proto, block.VegaTime, uint64(seqNum))
	if err != nil {
		t.Fatalf("failed to create trade from proto:%s", err)
	}
	return trade
}
