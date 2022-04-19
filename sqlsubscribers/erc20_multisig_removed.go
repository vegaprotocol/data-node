package sqlsubscribers

import (
	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/logging"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vega/events"
	"github.com/pkg/errors"
)

type ERC20MultiSigSignerRemovedEvent interface {
	events.Event
	Proto() eventspb.ERC20MultiSigSignerRemoved
}

//go:generate go run github.com/golang/mock/mockgen -destination mocks/withdrawals_mock.go -package mocks code.vegaprotocol.io/data-node/sqlsubscribers WithdrawalStore
type ERC20MultiSigSignerRemovedStore interface {
	Add(e *entities.ERC20MultiSigSignerRemoved) error
}

type ERC20MultiSigSignerRemoved struct {
	store ERC20MultiSigSignerRemovedStore
	log   *logging.Logger
}

func NewERC20MultiSigSignerRemoved(store ERC20MultiSigSignerRemovedStore, log *logging.Logger) *ERC20MultiSigSignerRemoved {
	return &ERC20MultiSigSignerRemoved{
		store: store,
		log:   log,
	}
}

func (t *ERC20MultiSigSignerRemoved) Types() []events.Type {
	return []events.Type{
		events.ERC20MultiSigSignerRemovedEvent,
	}
}

func (m *ERC20MultiSigSignerRemoved) Push(evt events.Event) error {
	switch e := evt.(type) {
	case TimeUpdateEvent:
		return nil // we have a timestamp in the event so we don't need to do anything here
	case ERC20MultiSigSignerRemovedEvent:
		return m.consume(e)
	default:
		return errors.Errorf("unknown event type %s", e.Type().String())
	}
}

func (m *ERC20MultiSigSignerRemoved) consume(event ERC20MultiSigSignerRemovedEvent) error {
	e := event.Proto()
	records, err := entities.ERC20MultiSigSignerRemovedFromProto(&e)
	if err != nil {
		return errors.Wrap(err, "converting signer-removed proto to database entity failed")
	}
	for _, r := range records {
		m.store.Add(r)
	}
	return nil
}
