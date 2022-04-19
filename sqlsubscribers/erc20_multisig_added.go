package sqlsubscribers

import (
	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/logging"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vega/events"
	"github.com/pkg/errors"
)

type ERC20MultiSigSignerAddedEvent interface {
	events.Event
	Proto() eventspb.ERC20MultiSigSignerAdded
}

//go:generate go run github.com/golang/mock/mockgen -destination mocks/withdrawals_mock.go -package mocks code.vegaprotocol.io/data-node/sqlsubscribers WithdrawalStore
type ERC20MultiSigSignerAddedStore interface {
	Add(e *entities.ERC20MultiSigSignerAdded) error
}

type ERC20MultiSigSignerAdded struct {
	store ERC20MultiSigSignerAddedStore
	log   *logging.Logger
}

func NewERC20MultiSigSignerAdded(store ERC20MultiSigSignerAddedStore, log *logging.Logger) *ERC20MultiSigSignerAdded {
	return &ERC20MultiSigSignerAdded{
		store: store,
		log:   log,
	}
}

func (t *ERC20MultiSigSignerAdded) Types() []events.Type {
	return []events.Type{
		events.ERC20MultiSigSignerAddedEvent,
	}
}

func (m *ERC20MultiSigSignerAdded) Push(evt events.Event) error {
	switch e := evt.(type) {
	case TimeUpdateEvent:
		return nil // we have a timestamp in the event so we don't need to do anything here
	case ERC20MultiSigSignerAddedEvent:
		return m.consume(e)
	default:
		return errors.Errorf("unknown event type %s", e.Type().String())
	}
}

func (m *ERC20MultiSigSignerAdded) consume(event ERC20MultiSigSignerAddedEvent) error {
	e := event.Proto()
	record, err := entities.ERC20MultiSigSignerAddedFromProto(&e)
	if err != nil {
		return errors.Wrap(err, "converting signer-added proto to database entity failed")
	}
	return m.store.Add(record)
}
