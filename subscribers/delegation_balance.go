package subscribers

import (
	"context"
	"strconv"

	"code.vegaprotocol.io/data-node/logging"
	types "code.vegaprotocol.io/protos/vega"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vega/events"
)

type DelegationBalanceEvent interface {
	events.Event
	Proto() eventspb.DelegationBalanceEvent
}

type PendingDelegationBalanceEvent interface {
	events.Event
	Proto() eventspb.PendingDelegationBalanceEvent
}

type DelegationBalanceSub struct {
	*Base

	epochStore EpochStore
	nodeStore  NodeStore

	log *logging.Logger
}

func NewDelegationBalanceSub(ctx context.Context, nodetore NodeStore, epochStore EpochStore, log *logging.Logger, ack bool) *DelegationBalanceSub {
	sub := &DelegationBalanceSub{
		Base:       NewBase(ctx, 10, ack),
		epochStore: epochStore,
		log:        log,
	}

	if sub.isRunning() {
		go sub.loop(ctx)
	}

	return sub
}

func (db *DelegationBalanceSub) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			db.Halt()
			return
		case e := <-db.ch:
			if db.isRunning() {
				db.Push(e...)
			}
		}
	}
}

func (db *DelegationBalanceSub) Push(evts ...events.Event) {
	if len(evts) == 0 {
		return
	}

	for _, e := range evts {
		switch et := e.(type) {
		case DelegationBalanceEvent:
			dbe := et.Proto()

			delegation := types.Delegation{
				EpochSeq: dbe.GetEpochSeq(),
				Party:    dbe.GetParty(),
				NodeId:   dbe.GetNodeId(),
				Amount:   strconv.FormatUint(dbe.GetAmount(), 10),
			}

			db.nodeStore.AddDelegation(delegation)
			db.epochStore.AddDelegation(delegation)

		case PendingDelegationBalanceEvent:
			continue
		default:
			db.log.Panic("Unknown event type in candles subscriber", logging.String("Type", et.Type().String()))
		}
	}
}

func (db *DelegationBalanceSub) Types() []events.Type {
	return []events.Type{
		events.DelegationBalanceEvent,
		events.PendingDelegationBalanceEvent,
	}
}