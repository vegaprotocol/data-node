package sqlsubscribers

import (
	"context"
	"time"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/data-node/subscribers"
	types "code.vegaprotocol.io/protos/vega"
	"code.vegaprotocol.io/vega/events"
)

type MarketDataEvent interface {
	events.Event
	MarketData() types.MarketData
}

//go:generate go run github.com/golang/mock/mockgen -destination mocks/market_data_mock.go -package mocks code.vegaprotocol.io/data-node/sqlsubscribers MarketDataStore
type MarketDataStore interface {
	Add(context.Context, *entities.MarketData) error
}

type MarketData struct {
	*subscribers.Base
	log        *logging.Logger
	store      MarketDataStore
	blockStore BlockStore
	dbTimeout  time.Duration
}

func NewMarketData(ctx context.Context, store MarketDataStore, blockStore BlockStore, log *logging.Logger, dbTimeout time.Duration) *MarketData {
	return &MarketData{
		Base:       subscribers.NewBase(ctx, 0, true),
		log:        log,
		store:      store,
		blockStore: blockStore,
		dbTimeout:  dbTimeout,
	}
}

func (md *MarketData) Types() []events.Type {
	return []events.Type{
		events.MarketDataEvent,
	}
}

func (md *MarketData) Push(events ...events.Event) {
	for _, e := range events {
		if data, ok := e.(MarketDataEvent); ok {
			md.consume(data)
		}
	}
}

func (md *MarketData) consume(event MarketDataEvent) {
	md.log.Debug("Received MarketData Event",
		logging.Int64("block", event.BlockNr()),
		logging.String("market", event.MarketData().Market),
	)

	block, err := md.blockStore.WaitForBlockHeight(event.BlockNr())
	if err != nil {
		md.log.Error("Can't add assert because we don't have block", logging.Error(err))
		return
	}

	ctx, cancel := context.WithTimeout(md.Base.Context(), md.dbTimeout)
	defer cancel()

	if err := md.addMarketData(ctx, event.MarketData(), block.VegaTime); err != nil {
		md.log.Error("Adding market data failed", logging.Error(err))
	}
}

func (md *MarketData) addMarketData(ctx context.Context, data types.MarketData, vegaTime time.Time) error {
	record, err := entities.MarketDataFromProto(data)
	if err != nil {
		return err
	}

	record.VegaTime = vegaTime

	return md.store.Add(ctx, record)
}
