package sqlsubscribers

import (
	"context"
	"time"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/protos/vega"
	"code.vegaprotocol.io/vega/events"
	"github.com/pkg/errors"
)

type LiquidityProvisionEvent interface {
	events.Event
	LiquidityProvision() *vega.LiquidityProvision
}

//go:generate go run github.com/golang/mock/mockgen -destination mocks/liquidity_provision_mock.go -package mocks code.vegaprotocol.io/data-node/sqlsubscribers LiquidityProvisionStore
type LiquidityProvisionStore interface {
	Upsert(context.Context, *entities.LiquidityProvision) error
}

type LiquidityProvision struct {
	store    LiquidityProvisionStore
	log      *logging.Logger
	vegaTime time.Time
}

func NewLiquidityProvision(store LiquidityProvisionStore, log *logging.Logger) *LiquidityProvision {
	return &LiquidityProvision{
		store: store,
		log:   log,
	}
}

func (lp *LiquidityProvision) Types() []events.Type {
	return []events.Type{events.LiquidityProvisionEvent}
}

func (lp *LiquidityProvision) Push(ctx context.Context, evt events.Event) error {
	switch e := evt.(type) {
	case TimeUpdateEvent:
		lp.vegaTime = e.Time()
	case LiquidityProvisionEvent:
		return lp.consume(ctx, e)
	default:
		return errors.Errorf("unknown event type %s", e.Type().String())
	}

	return nil
}

func (lp *LiquidityProvision) consume(ctx context.Context, event LiquidityProvisionEvent) error {
	provision := event.LiquidityProvision()
	entity, err := entities.LiquidityProvisionFromProto(provision, lp.vegaTime)
	if err != nil {
		return errors.Wrap(err, "converting liquidity provision to database entity failed")
	}

	return errors.Wrap(lp.store.Upsert(ctx, entity), "inserting liquidity provision to SQL store failed")
}
