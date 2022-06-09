package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/data-node/utils"
	"golang.org/x/exp/maps"
)

//go:generate go run github.com/golang/mock/mockgen -destination mocks/market_data_mock.go -package mocks code.vegaprotocol.io/data-node/service MarketDataStore
type MarketDataStore interface {
	Add(data *entities.MarketData) error
	Flush(ctx context.Context) ([]*entities.MarketData, error)
	GetMarketDataByID(ctx context.Context, marketID string) (entities.MarketData, error)
	GetMarketsData(ctx context.Context) ([]entities.MarketData, error)
	GetBetweenDatesByID(ctx context.Context, marketID string, start, end time.Time, pagination entities.OffsetPagination) ([]entities.MarketData, error)
	GetFromDateByID(ctx context.Context, marketID string, start time.Time, pagination entities.OffsetPagination) ([]entities.MarketData, error)
	GetToDateByID(ctx context.Context, marketID string, end time.Time, pagination entities.OffsetPagination) ([]entities.MarketData, error)
}

type MarketData struct {
	store     MarketDataStore
	log       *logging.Logger
	observer  utils.Observer[*entities.MarketData]
	cache     map[entities.MarketID]*entities.MarketData
	cacheLock sync.RWMutex
}

func NewMarketData(store MarketDataStore, log *logging.Logger) *MarketData {
	return &MarketData{
		log:      log,
		store:    store,
		observer: utils.NewObserver[*entities.MarketData]("market_data", log, 0, 0),
		cache:    make(map[entities.MarketID]*entities.MarketData),
	}
}

func (m *MarketData) Add(data *entities.MarketData) error {
	if err := m.store.Add(data); err != nil {
		return err
	}
	m.cacheLock.Lock()
	m.cache[data.Market] = data
	m.cacheLock.Unlock()
	return nil
}

func (m *MarketData) Flush(ctx context.Context) error {
	flushed, err := m.store.Flush(ctx)
	if err != nil {
		return err
	}
	m.observer.Notify(flushed)
	return nil
}

func (m *MarketData) Initialise(ctx context.Context) error {
	m.cacheLock.Lock()
	defer m.cacheLock.Unlock()

	all, err := m.store.GetMarketsData(ctx)
	if err != nil {
		return err
	}
	for i := 0; i < len(all); i++ {
		m.cache[all[i].Market] = &all[i]
	}
	return nil
}

func (m *MarketData) GetMarketDataByID(ctx context.Context, marketID string) (entities.MarketData, error) {
	m.cacheLock.RLock()
	defer m.cacheLock.RUnlock()

	data, ok := m.cache[entities.NewMarketID(marketID)]
	if !ok {
		return entities.MarketData{}, fmt.Errorf("no market data for market: %v", marketID)
	}
	return *data, nil
}

func (m *MarketData) GetMarketsData(ctx context.Context) ([]entities.MarketData, error) {
	m.cacheLock.RLock()
	defer m.cacheLock.RUnlock()

	pData := maps.Values(m.cache)
	data := make([]entities.MarketData, len(pData))
	for i := 0; i < len(pData); i++ {
		data[i] = *pData[i]
	}
	return data, nil
}

func (m *MarketData) GetBetweenDatesByID(ctx context.Context, marketID string, start, end time.Time, pagination entities.OffsetPagination) ([]entities.MarketData, error) {
	return m.store.GetBetweenDatesByID(ctx, marketID, start, end, pagination)
}

func (m *MarketData) GetFromDateByID(ctx context.Context, marketID string, start time.Time, pagination entities.OffsetPagination) ([]entities.MarketData, error) {
	return m.store.GetFromDateByID(ctx, marketID, start, pagination)
}

func (m *MarketData) GetToDateByID(ctx context.Context, marketID string, end time.Time, pagination entities.OffsetPagination) ([]entities.MarketData, error) {
	return m.store.GetToDateByID(ctx, marketID, end, pagination)
}

func (m *MarketData) ObserveMarketData(
	ctx context.Context, retries int, marketID string,
) (<-chan []*entities.MarketData, uint64) {
	ch, ref := m.observer.Observe(ctx,
		retries,
		func(md *entities.MarketData) bool { return len(marketID) == 0 || marketID == md.Market.String() })
	return ch, ref
}
