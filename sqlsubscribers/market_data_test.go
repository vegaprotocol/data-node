package sqlsubscribers

import (
	"context"
	"errors"
	"testing"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/data-node/sqlsubscribers/mocks"
	"code.vegaprotocol.io/vega/events"
	"code.vegaprotocol.io/vega/types"
	"github.com/golang/mock/gomock"
)

func Test_MarketData_Push(t *testing.T) {
	t.Run("Should call market data store Add if Block Height is received", testShouldCallStoreAddIfBlockHeightIsReceived)
	t.Run("Should not call market data store Add if block height is not received", testShouldNotCallStoreAddIfBlockHeightNotReceived)
}

func testShouldCallStoreAddIfBlockHeightIsReceived(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockMarketDataStore(ctrl)
	blockStore := mocks.NewMockBlockStore(ctrl)

	store.EXPECT().Add(gomock.Any()).Times(1)
	blockStore.EXPECT().WaitForBlockHeight(gomock.Any()).Times(1)

	subscriber := NewMarketData(context.Background(), logging.NewTestLogger(), store, blockStore)
	subscriber.Push(events.NewMarketDataEvent(context.Background(), types.MarketData{}))
}

func testShouldNotCallStoreAddIfBlockHeightNotReceived(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := mocks.NewMockMarketDataStore(ctrl)
	blockStore := mocks.NewMockBlockStore(ctrl)

	store.EXPECT().Add(gomock.Any()).Times(0)
	blockStore.EXPECT().WaitForBlockHeight(int64(0)).DoAndReturn(func(int64) (entities.Block, error) {
		return entities.Block{}, errors.New("some error")
	})

	subscriber := NewMarketData(context.Background(), logging.NewTestLogger(), store, blockStore)
	subscriber.Push(events.NewMarketDataEvent(context.Background(), types.MarketData{}))
}
