package api_test

import (
	"context"
	"sync"

	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/data-node/subscribers"
	"code.vegaprotocol.io/vega/events"
)

type EventSubscriber struct {
	*subscribers.Base
	events chan events.Event

	closed bool
	mu     sync.RWMutex
	log    *logging.Logger
}

func NewEventSubscriber(ctx context.Context, log *logging.Logger) *EventSubscriber {
	t := &EventSubscriber{
		Base:   subscribers.NewBase(ctx, 10, true),
		events: make(chan events.Event, 20),
		log:    log,
	}

	return t
}

func (t *EventSubscriber) Push(evts ...events.Event) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.log.Debug("TIME SUB DONE")
	if t.closed {
		return
	}

	for _, e := range evts {
		t.events <- e
	}
}

func (t *EventSubscriber) ReceivedEvent(ctx context.Context) (events.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case e := <-t.events:
		return e, nil
	}
}

func (t *EventSubscriber) Halt() {
	t.mu.Lock()
	close(t.events)
	t.closed = true
	t.mu.Unlock()
}

func (*EventSubscriber) Types() []events.Type {
	return []events.Type{}
}
