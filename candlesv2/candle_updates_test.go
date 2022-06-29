// Copyright (c) 2022 Gobalsky Labs Limited
//
// Use of this software is governed by the Business Source License included
// in the LICENSE file and at https://www.mariadb.com/bsl11.
//
// Change Date: 18 months from the later of the date of the first publicly
// available Distribution of this version of the repository, and 25 June 2022.
//
// On the date above, in accordance with the Business Source License, use
// of this software will be governed by version 3 or later of the GNU General
// Public License.

package candlesv2_test

import (
	"context"
	"testing"
	"time"

	"code.vegaprotocol.io/data-node/candlesv2"
	"code.vegaprotocol.io/data-node/entities"
	v2 "code.vegaprotocol.io/protos/data-node/api/v2"

	"github.com/shopspring/decimal"

	"github.com/stretchr/testify/assert"

	"code.vegaprotocol.io/data-node/config/encoding"
	"code.vegaprotocol.io/data-node/logging"
)

type testCandleSource struct {
	candles chan []entities.Candle
}

func (t *testCandleSource) GetCandleDataForTimeSpan(ctx context.Context, candleId string, from *time.Time, to *time.Time,
	p entities.CursorPagination,
) entities.ConnectionData[*v2.CandleEdge, entities.Candle] {
	select {
	case c := <-t.candles:
		return entities.ConnectionData[*v2.CandleEdge, entities.Candle]{
			TotalCount: 0,
			Entities:   c,
			PageInfo:   entities.PageInfo{},
			Err:        nil,
		}
	default:
		return entities.ConnectionData[*v2.CandleEdge, entities.Candle]{
			TotalCount: 0,
			Entities:   nil,
			PageInfo:   entities.PageInfo{},
			Err:        nil,
		}
	}
}

func TestSubscribe(t *testing.T) {
	testCandleSource := &testCandleSource{candles: make(chan []entities.Candle)}

	updates, _ := candlesv2.NewCandleUpdates(context.Background(), logging.NewTestLogger(), "testCandles",
		testCandleSource, newTestCandleConfig(1).CandleUpdates)
	startTime := time.Now()

	_, out1 := updates.Subscribe()
	_, out2 := updates.Subscribe()

	expectedCandle := createCandle(startTime, startTime, 1, 1, 1, 1, 10)
	testCandleSource.candles <- []entities.Candle{expectedCandle}

	candle1 := <-out1
	assert.Equal(t, expectedCandle, candle1)

	candle2 := <-out2
	assert.Equal(t, expectedCandle, candle2)

	expectedCandle = createCandle(startTime.Add(1*time.Minute), startTime.Add(1*time.Minute), 2, 2, 2, 2, 20)
	testCandleSource.candles <- []entities.Candle{expectedCandle}

	candle1 = <-out1
	assert.Equal(t, expectedCandle, candle1)

	candle2 = <-out2
	assert.Equal(t, expectedCandle, candle2)
}

func TestUnsubscribe(t *testing.T) {
	testCandleSource := &testCandleSource{candles: make(chan []entities.Candle)}

	updates, _ := candlesv2.NewCandleUpdates(context.Background(), logging.NewTestLogger(), "testCandles",
		testCandleSource, newTestCandleConfig(1).CandleUpdates)
	startTime := time.Now()

	id, out1 := updates.Subscribe()

	expectedCandle := createCandle(startTime, startTime, 1, 1, 1, 1, 10)
	testCandleSource.candles <- []entities.Candle{expectedCandle}

	candle1 := <-out1
	assert.Equal(t, expectedCandle, candle1)

	updates.Unsubscribe(id)

	_, ok := <-out1
	assert.False(t, ok, "candle should be closed")
}

func TestNewSubscriberAlwaysGetsLastCandle(t *testing.T) {
	testCandleSource := &testCandleSource{candles: make(chan []entities.Candle)}

	updates, _ := candlesv2.NewCandleUpdates(context.Background(), logging.NewTestLogger(), "testCandles",
		testCandleSource, newTestCandleConfig(1).CandleUpdates)
	startTime := time.Now()

	_, out1 := updates.Subscribe()

	expectedCandle := createCandle(startTime, startTime, 1, 1, 1, 1, 10)
	testCandleSource.candles <- []entities.Candle{expectedCandle}

	candle1 := <-out1
	assert.Equal(t, expectedCandle, candle1)

	_, out2 := updates.Subscribe()
	candle2 := <-out2
	assert.Equal(t, expectedCandle, candle2)
}

func TestSlowConsumersChannelIsClosed(t *testing.T) {
	testCandleSource := &testCandleSource{candles: make(chan []entities.Candle)}

	updates, _ := candlesv2.NewCandleUpdates(context.Background(), logging.NewTestLogger(), "testCandles",
		testCandleSource, newTestCandleConfig(1).CandleUpdates)
	startTime := time.Now()

	_, out1 := updates.Subscribe()

	expectedCandle := createCandle(startTime, startTime, 1, 1, 1, 1, 10)
	candle2 := createCandle(startTime.Add(1*time.Minute), startTime.Add(1*time.Minute), 2, 2, 2, 2, 20)
	testCandleSource.candles <- []entities.Candle{expectedCandle}
	testCandleSource.candles <- []entities.Candle{candle2}

	candle1 := <-out1
	assert.Equal(t, expectedCandle, candle1)

	_, ok := <-out1
	assert.False(t, ok, "channel should be closed")
}

func newTestCandleConfig(bufferSize int) candlesv2.Config {
	conf := candlesv2.NewDefaultConfig()
	conf.CandleUpdates = candlesv2.CandleUpdatesConfig{
		CandleUpdatesStreamBufferSize: bufferSize,
		CandleUpdatesStreamInterval:   encoding.Duration{Duration: time.Duration(1 * time.Microsecond)},
		CandlesFetchTimeout:           encoding.Duration{Duration: time.Duration(2 * time.Minute)},
	}

	return conf
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
