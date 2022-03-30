package sqlsubscribers

import (
	"context"
	"errors"
	"sync"
	"time"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/data-node/sqlstore"
	"code.vegaprotocol.io/vega/events"
	"code.vegaprotocol.io/vega/types/num"
)

type positionEventBase interface {
	events.Event
	PartyID() string
	MarketID() string
	Timestamp() int64
}

type positionSettlement interface {
	positionEventBase
	Price() *num.Uint
	PositionFactor() num.Decimal
	Trades() []events.TradeSettlement
}

type lossSocialization interface {
	positionEventBase
	Amount() *num.Int
}

type settleDistressed interface {
	positionEventBase
	Margin() *num.Uint
}

type positionState interface {
	positionEventBase
	Size() int64
}

//go:generate go run github.com/golang/mock/mockgen -destination mocks/positions_mock.go -package mocks code.vegaprotocol.io/data-node/sqlsubscribers PositionStore
type PositionStore interface {
	Add(context.Context, entities.Position) error
	GetByMarket(ctx context.Context, marketID entities.MarketID) ([]entities.Position, error)
	GetByMarketAndParty(ctx context.Context, marketID entities.MarketID, partyID entities.PartyID) (entities.Position, error)
}

type Position struct {
	store    PositionStore
	log      *logging.Logger
	vegaTime time.Time
	mutex    sync.Mutex
}

func NewPosition(
	store PositionStore,
	log *logging.Logger,
) *Position {
	t := &Position{
		store: store,
		log:   log,
	}
	return t
}

func (t *Position) Types() []events.Type {
	return []events.Type{
		events.SettlePositionEvent,
		events.SettleDistressedEvent,
		events.LossSocializationEvent,
	}
}

func (nl *Position) Push(evt events.Event) {
	switch event := evt.(type) {
	case TimeUpdateEvent:
		nl.vegaTime = event.Time()
	case positionSettlement:
		nl.handlePositionSettlement(event)
	case lossSocialization:
		nl.handleLossSocialization(event)
	case settleDistressed:
		nl.handleSettleDestressed(event)
	case positionState:
		nl.handlePositionState(event)
	default:
		nl.log.Panic("Unknown event type in position subscriber",
			logging.String("Type", event.Type().String()))
	}
}

func (ps *Position) handlePositionSettlement(event positionSettlement) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	pos := ps.getPosition(event)
	pos.UpdateWithPositionSettlement(event)
	ps.updatePosition(pos)
}

func (ps *Position) handleLossSocialization(event lossSocialization) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	pos := ps.getPosition(event)
	pos.UpdateWithLossSocialization(event)
	ps.updatePosition(pos)
}

func (ps *Position) handleSettleDestressed(event settleDistressed) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	pos := ps.getPosition(event)
	pos.UpdateWithSettleDestressed(event)
	ps.updatePosition(pos)
}

func (ps *Position) handlePositionState(event positionState) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	pos := ps.getPosition(event)
	pos.UpdateWithPositionState(event)
	ps.updatePosition(pos)
}

func (ps *Position) getPosition(e positionEventBase) entities.Position {
	mID := entities.NewMarketID(e.MarketID())
	pID := entities.NewPartyID(e.PartyID())

	position, err := ps.store.GetByMarketAndParty(context.Background(), mID, pID)
	if errors.Is(err, sqlstore.ErrPositionNotFound) {
		return entities.NewEmptyPosition(mID, pID)
	}

	if err != nil {
		// TODO: Can we do something less drastic here? If we can't get existing positions
		//       things are a bit screwed as we'll start writing down wrong aggregates.
		panic("unable to query for existing position")
	}

	return position
}

func (ps *Position) updatePosition(pos entities.Position) {
	pos.VegaTime = ps.vegaTime

	err := ps.store.Add(context.Background(), pos)
	if err != nil {
		ps.log.Error("Error updating position", logging.Error(err))
	}
}
