package sqlstore_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/sqlstore"
)

func Test_CandleUpdatesOrder(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore, err := sqlstore.NewCandles(context.Background(), testStore, "trades", newTestCandleConfig(5))
	if err != nil {
		t.Fatalf("creating candle store:%s", err)
	}
	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	insertTestTrade(t, tradeStore, 1, 10, block, 0)

	_, candleId, _ := candleStore.GetCandleIdForIntervalAndGroup(context.Background(), "1 Minute", testMarket)
	_, out, err := tradeStore.SubscribeToTradesCandle(context.Background(), candleId)
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

func Test_CandleUpdatesSlowConsumer(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore, err := sqlstore.NewCandles(context.Background(), testStore, "trades", newTestCandleConfig(1))
	if err != nil {
		t.Fatalf("creating candle store:%s", err)
	}

	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	insertTestTrade(t, tradeStore, 1, 10, block, 0)

	_, candleId, _ := candleStore.GetCandleIdForIntervalAndGroup(context.Background(), "1 Minute", testMarket)
	_, out, err := tradeStore.SubscribeToTradesCandle(context.Background(), candleId)
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

func Test_CandleUpdatesSubscribe(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore, err := sqlstore.NewCandles(context.Background(), testStore, "trades", newTestCandleConfig(1))
	if err != nil {
		t.Fatalf("creating candle store:%s", err)
	}
	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	insertTestTrade(t, tradeStore, 1, 10, block, 0)

	_, candleId, _ := candleStore.GetCandleIdForIntervalAndGroup(context.Background(), "1 Minute", testMarket)
	_, out1, err := tradeStore.SubscribeToTradesCandle(context.Background(), candleId)
	if err != nil {
		t.Fatalf("failed to subscribe: %s", err)
	}
	_, out2, err := tradeStore.SubscribeToTradesCandle(context.Background(), candleId)
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

func Test_CandleUpdatesSubscribeWhenNoCandleExistsInStartPeriod(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore, err := sqlstore.NewCandles(context.Background(), testStore, "trades", newTestCandleConfig(1))
	if err != nil {
		t.Fatalf("creating candle store:%s", err)
	}
	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	insertTestTrade(t, tradeStore, 1, 10, block, 0)

	_, candleId, _ := candleStore.GetCandleIdForIntervalAndGroup(context.Background(), "1 Minute", testMarket)
	_, out, err := tradeStore.SubscribeToTradesCandle(context.Background(), candleId)
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

func Test_CandleUpdatesTradesConflatedIntoCandle(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore, err := sqlstore.NewCandles(context.Background(), testStore, "trades", newTestCandleConfig(1))
	if err != nil {
		t.Fatalf("creating candle store:%s", err)
	}
	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	insertTestTrade(t, tradeStore, 1, 10, block, 0)

	_, candleId, _ := candleStore.GetCandleIdForIntervalAndGroup(context.Background(), "1 Minute", testMarket)
	_, out, err := tradeStore.SubscribeToTradesCandle(context.Background(), candleId)
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

func Test_CandleUpdatesUnsubscribe(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore, err := sqlstore.NewCandles(context.Background(), testStore, "trades", newTestCandleConfig(1))
	if err != nil {
		t.Fatalf("creating candle store:%s", err)
	}
	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	insertTestTrade(t, tradeStore, 1, 10, block, 0)

	_, candleId, _ := candleStore.GetCandleIdForIntervalAndGroup(context.Background(), "1 Minute", testMarket)
	subscriptionId, out, err := tradeStore.SubscribeToTradesCandle(context.Background(), candleId)
	if err != nil {
		t.Fatalf("failed to subscribe: %s", err)
	}

	insertTestTrade(t, tradeStore, 2, 20, block, 1)

	candle1 := <-out
	expectedCandle := createCandle(startTime, startTime.Add(1*time.Microsecond), 1, 2, 2, 1, 30)
	assert.Equal(t, expectedCandle, candle1)

	tradeStore.UnsubscribeFromTradesCandle(subscriptionId)

	block = nextBlock(t, bs, block, 1*time.Second)
	insertTestTrade(t, tradeStore, 4, 20, block, 0)

	select {
	case _, ok := <-out:
		if ok {
			t.Fatalf("channel should be closed")
		}
	default:
	}

	block = nextBlock(t, bs, block, 1*time.Second)
	insertTestTrade(t, tradeStore, 5, 20, block, 0)

	subscriptionId2, out, err := tradeStore.SubscribeToTradesCandle(context.Background(), candleId)
	if err != nil {
		t.Fatalf("failed to subscribe: %s", err)
	}

	assert.NotEqual(t, subscriptionId2, subscriptionId)

	block = nextBlock(t, bs, block, 1*time.Second)
	insertTestTrade(t, tradeStore, 6, 20, block, 0)

	candle2 := <-out
	expectedCandle = createCandle(startTime, block.VegaTime, 1, 6, 6, 1, 90)
	assert.Equal(t, expectedCandle, candle2)

}

func Test_CandleUpdatesOverPeriodBoundary(t *testing.T) {
	defer testStore.DeleteEverything()

	candleStore, err := sqlstore.NewCandles(context.Background(), testStore, "trades", newTestCandleConfig(1))
	if err != nil {
		t.Fatalf("creating candle store:%s", err)
	}
	tradeStore := sqlstore.NewTrades(testStore, candleStore)
	bs := sqlstore.NewBlocks(testStore)

	startTime := time.Unix(Midnight3rdMarch2022InUnixTime, 0)
	block := addTestBlockForTime(t, bs, startTime)

	insertTestTrade(t, tradeStore, 1, 10, block, 0)

	_, candleId, _ := candleStore.GetCandleIdForIntervalAndGroup(context.Background(), "1 Minute", testMarket)
	_, out, err := tradeStore.SubscribeToTradesCandle(context.Background(), candleId)
	if err != nil {
		t.Fatalf("failed to subscribe to candle:%s", err)
	}

	time.Sleep(1 * time.Second)
	insertTestTrade(t, tradeStore, 1, 10, block, 1)

	candle1 := <-out
	expectedCandle := createCandle(startTime, startTime.Add(1*time.Microsecond), 1, 1, 1, 1, 20)
	assert.Equal(t, expectedCandle, candle1)

	block = nextBlock(t, bs, block, 1*time.Second)
	insertTestTrade(t, tradeStore, 2, 20, block, 2)

	candle2 := <-out
	expectedCandle = createCandle(startTime, startTime.Add(1*time.Second).Add(2*time.Microsecond), 1, 2, 2, 1, 40)

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
