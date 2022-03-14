package sqlstore_test

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/sqlstore"
)

func Test_CandleEventStreamOrder(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore := sqlstore.NewCandles(testStore, newTestCandleConfig(5))
	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	insertTestTrade(t, tradeStore, 1, 10, block, 0)

	testMarketId, _ := hex.DecodeString(testMarket)
	_, out, err := tradeStore.SubscribeToCandle(context.Background(), testMarketId, secondsToInterval(60))
	if err != nil {
		t.Fatalf("failed to subscribe: %s", err)
	}

	insertTestTrade(t, tradeStore, 2, 20, block, 1)
	insertTestTrade(t, tradeStore, 3, 20, block, 2)
	insertTestTrade(t, tradeStore, 4, 20, block, 3)

	candle1 := <-out
	expectedCandle := createCandle(startTime, startTime.Add(1*time.Microsecond), 1, 2, 2, 1, 30)
	assert.Equal(t, expectedCandle, candle1)

	candle2 := <-out
	expectedCandle = createCandle(startTime, startTime.Add(2*time.Microsecond), 1, 3, 3, 1, 50)
	assert.Equal(t, expectedCandle, candle2)

	candle3 := <-out
	expectedCandle = createCandle(startTime, startTime.Add(3*time.Microsecond), 1, 4, 4, 1, 70)
	assert.Equal(t, expectedCandle, candle3)
}

func Test_CandleEventSlowConsumer(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore := sqlstore.NewCandles(testStore, newTestCandleConfig(1))
	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	insertTestTrade(t, tradeStore, 1, 10, block, 0)

	testMarketId, _ := hex.DecodeString(testMarket)
	_, out, err := tradeStore.SubscribeToCandle(context.Background(), testMarketId, secondsToInterval(60))
	if err != nil {
		t.Fatalf("failed to subscribe: %s", err)
	}

	insertTestTrade(t, tradeStore, 2, 20, block, 1)
	insertTestTrade(t, tradeStore, 3, 20, block, 2)

	candle1 := <-out
	expectedCandle := createCandle(startTime, startTime.Add(1*time.Microsecond), 1, 2, 2, 1, 30)
	assert.Equal(t, expectedCandle, candle1)

	select {
	case _, ok := <-out:
		if ok {
			t.Fatalf("channel should be closed")
		}
	default:
	}
}

func Test_CandleEventStreamSubscribe(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore := sqlstore.NewCandles(testStore, newTestCandleConfig(1))
	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	insertTestTrade(t, tradeStore, 1, 10, block, 0)

	testMarketId, _ := hex.DecodeString(testMarket)
	_, out1, err := tradeStore.SubscribeToCandle(context.Background(), testMarketId, secondsToInterval(60))
	if err != nil {
		t.Fatalf("failed to subscribe: %s", err)
	}
	_, out2, err := tradeStore.SubscribeToCandle(context.Background(), testMarketId, secondsToInterval(60))
	if err != nil {
		t.Fatalf("failed to subscribe: %s", err)
	}

	block = nextBlock(t, bs, block, 1*time.Second)
	insertTestTrade(t, tradeStore, 2, 20, block, 1)

	candle1 := <-out1
	candle2 := <-out2
	expectedCandle := createCandle(startTime, startTime.Add(1*time.Second).Add(1*time.Microsecond),
		1, 2, 2, 1, 30)
	assert.Equal(t, expectedCandle, candle1)
	assert.Equal(t, expectedCandle, candle2)
}

func Test_CandleEventStreamSubscribeWhenNoCandleExistsInStartPeriod(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore := sqlstore.NewCandles(testStore, newTestCandleConfig(1))
	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	insertTestTrade(t, tradeStore, 1, 10, block, 0)

	testMarketId, _ := hex.DecodeString(testMarket)
	_, out, err := tradeStore.SubscribeToCandle(context.Background(), testMarketId, secondsToInterval(60))
	if err != nil {
		t.Fatalf("failed to subscribe: %s", err)
	}

	block = nextBlock(t, bs, block, 1*time.Hour)
	insertTestTrade(t, tradeStore, 2, 20, block, 1)

	candle1 := <-out
	expectedCandle := createCandle(startTime.Add(1*time.Hour), startTime.Add(1*time.Hour).Add(1*time.Microsecond),
		2, 2, 2, 2, 20)
	assert.Equal(t, expectedCandle, candle1)
}

func Test_CandleEventStreamTradesConflatedIntoCandle(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore := sqlstore.NewCandles(testStore, newTestCandleConfig(1))
	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	insertTestTrade(t, tradeStore, 1, 10, block, 0)

	testMarketId, _ := hex.DecodeString(testMarket)
	_, out, err := tradeStore.SubscribeToCandle(context.Background(), testMarketId, secondsToInterval(60))
	if err != nil {
		t.Fatalf("failed to subscribe: %s", err)
	}

	block = nextBlock(t, bs, block, 1*time.Minute)
	tradeStore.Add(createTestTrade(t, 2, 20, block, 1))
	tradeStore.Add(createTestTrade(t, 3, 20, block, 2))
	tradeStore.Add(createTestTrade(t, 4, 20, block, 3))
	tradeStore.OnTimeUpdateEvent(context.Background())

	candle1 := <-out
	expectedCandle := createCandle(startTime.Add(1*time.Minute), startTime.Add(1*time.Minute).Add(3*time.Microsecond), 2, 4, 4, 2, 60)
	assert.Equal(t, expectedCandle, candle1)
}

func nextBlock(t *testing.T, bs *sqlstore.Blocks, block entities.Block, duration time.Duration) entities.Block {
	return addTestBlockForTime(t, bs, block.VegaTime.Add(duration))
}

func Test_CandleEventStreamUnsubscribe(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore := sqlstore.NewCandles(testStore, newTestCandleConfig(1))
	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	insertTestTrade(t, tradeStore, 1, 10, block, 0)

	testMarketId, _ := hex.DecodeString(testMarket)
	subscriberId, out, err := tradeStore.SubscribeToCandle(context.Background(), testMarketId, secondsToInterval(60))
	if err != nil {
		t.Fatalf("failed to subscribe: %s", err)
	}

	insertTestTrade(t, tradeStore, 2, 20, block, 1)

	candle1 := <-out
	expectedCandle := createCandle(startTime, startTime.Add(1*time.Microsecond), 1, 2, 2, 1, 30)
	assert.Equal(t, expectedCandle, candle1)

	tradeStore.UnsubscribeFromCandle(subscriberId)

	block = nextBlock(t, bs, block, 1*time.Second)
	insertTestTrade(t, tradeStore, 4, 20, block, 0)

	select {
	case _, ok := <-out:
		if ok {
			t.Fatalf("channel should be closed")
		}
	default:
	}
}

func Test_CandleEventStreamOverPeriodBoundary(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore := sqlstore.NewCandles(testStore, newTestCandleConfig(1))
	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	interval := secondsToInterval(60)
	testMarketId, _ := hex.DecodeString(testMarket)
	_, out, _ := tradeStore.SubscribeToCandle(context.Background(), testMarketId, interval)
	insertTestTrade(t, tradeStore, 1, 10, block, 0)

	candle1 := <-out
	expectedCandle := createCandle(startTime, startTime, 1, 1, 1, 1, 10)
	assert.Equal(t, expectedCandle, candle1)

	block = nextBlock(t, bs, block, 1*time.Second)
	insertTestTrade(t, tradeStore, 2, 20, block, 1)

	candle2 := <-out
	expectedCandle = createCandle(startTime, startTime.Add(1*time.Second).Add(1*time.Microsecond), 1, 2, 2, 1, 30)

	assert.Equal(t, expectedCandle, candle2)

	block = nextBlock(t, bs, block, 1*time.Minute)
	insertTestTrade(t, tradeStore, 3, 20, block, 0)

	candle3 := <-out
	expectedCandle = createCandle(startTime.Add(1*time.Minute), startTime.Add(1*time.Second).Add(1*time.Minute), 3, 3, 3, 3, 20)
	assert.Equal(t, expectedCandle, candle3)

	block = nextBlock(t, bs, block, 1*time.Second)
	insertTestTrade(t, tradeStore, 4, 20, block, 0)

	candle4 := <-out
	expectedCandle = createCandle(startTime.Add(1*time.Minute), startTime.Add(2*time.Second).Add(1*time.Minute), 3, 4, 4, 3, 40)
	assert.Equal(t, expectedCandle, candle4)

	block = nextBlock(t, bs, block, 1*time.Hour)
	insertTestTrade(t, tradeStore, 5, 20, block, 0)

	candle5 := <-out
	expectedCandle = createCandle(startTime.Add(1*time.Minute).Add(1*time.Hour), startTime.Add(2*time.Second).Add(1*time.Minute).Add(1*time.Hour),
		5, 5, 5, 5, 20)
	assert.Equal(t, expectedCandle, candle5)
}
