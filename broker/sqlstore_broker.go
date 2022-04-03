package broker

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"code.vegaprotocol.io/data-node/logging"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vega/events"
)

type SqlBrokerSubscriber interface {
	Push(val events.Event)
	Types() []events.Type
}

type SqlStoreEventBroker interface {
	Receive(ctx context.Context) error
}

func NewSqlStoreBroker(log *logging.Logger, config Config, chainInfo ChainInfoI,
	eventsource eventSource, eventTypeBufferSize int, subs ...SqlBrokerSubscriber,
) SqlStoreEventBroker {
	log = log.Named(namedLogger)
	log.SetLevel(config.Level.Get())
	if config.UseSequentialSqlStoreBroker {
		return newSequentialSqlStoreBroker(log, chainInfo, eventsource, eventTypeBufferSize, subs)
	} else {
		return newConcurrentSqlStoreBroker(log, chainInfo, eventsource, eventTypeBufferSize, subs)
	}
}

// concurrentSqlStoreBroker : push events to each subscriber with a go-routine per type
type concurrentSqlStoreBroker struct {
	typeToSubs          map[events.Type][]SqlBrokerSubscriber
	typeToEvtCh         map[events.Type]chan events.Event
	eventSource         eventSource
	chainInfo           ChainInfoI
	eventTypeBufferSize int
	log                 *logging.Logger
}

func newConcurrentSqlStoreBroker(log *logging.Logger, chainInfo ChainInfoI, eventsource eventSource, eventTypeBufferSize int,
	subs []SqlBrokerSubscriber,
) *concurrentSqlStoreBroker {
	b := &concurrentSqlStoreBroker{
		typeToSubs:          map[events.Type][]SqlBrokerSubscriber{},
		typeToEvtCh:         map[events.Type]chan events.Event{},
		eventSource:         eventsource,
		chainInfo:           chainInfo,
		eventTypeBufferSize: eventTypeBufferSize,
		log:                 log,
	}

	for _, s := range subs {
		b.subscribe(s)
	}
	return b
}

func (b *concurrentSqlStoreBroker) Receive(ctx context.Context) error {
	if err := b.eventSource.Listen(); err != nil {
		return err
	}

	receiveCh, errCh := b.eventSource.Receive(ctx)
	b.startSendingEvents(ctx)

	for e := range receiveCh {
		if err := checkChainID(b.chainInfo, e.ChainID()); err != nil {
			return err
		}

		// If the event is a time event send it to all type channels, this indicates a new block start (for now)
		if e.Type() == events.TimeUpdate {
			for _, ch := range b.typeToEvtCh {
				ch <- e
			}

			waitGroup := &WaitGroupEvent{}
			waitGroup.Add(len(b.typeToEvtCh))

			for _, ch := range b.typeToEvtCh {
				ch <- waitGroup
			}

			waitGroup.Wait()

		} else {
			if ch, ok := b.typeToEvtCh[e.Type()]; ok {
				ch <- e
			}
		}
	}

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func (b *concurrentSqlStoreBroker) subscribe(s SqlBrokerSubscriber) {
	for _, evtType := range s.Types() {
		if _, exists := b.typeToEvtCh[evtType]; !exists {
			ch := make(chan events.Event, b.eventTypeBufferSize)
			b.typeToEvtCh[evtType] = ch
		}

		b.typeToSubs[evtType] = append(b.typeToSubs[evtType], s)
	}
}

func (b *concurrentSqlStoreBroker) startSendingEvents(ctx context.Context) {
	for t, ch := range b.typeToEvtCh {
		go func(ch chan events.Event, subs []SqlBrokerSubscriber) {
			for {
				select {
				case <-ctx.Done():
					return
				case evt := <-ch:
					if evt.Type() == WaitGroupEventType {
						waitGroup := evt.(*WaitGroupEvent)
						waitGroup.Done()
					} else if evt.Type() == events.TimeUpdate {
						time := evt.(*events.Time)
						for _, sub := range subs {
							sub.Push(time)
						}
					} else {
						for _, sub := range subs {
							select {
							case <-ctx.Done():
								return
							default:
								sub.Push(evt)
							}
						}
					}
				}
			}
		}(ch, b.typeToSubs[t])
	}
}

const WaitGroupEventType = events.Type(math.MaxInt32)

type WaitGroupEvent struct {
	sync.WaitGroup
}

func (c *WaitGroupEvent) Type() events.Type {
	return WaitGroupEventType
}

func (c *WaitGroupEvent) Context() context.Context {
	panic("implement me")
}

func (c *WaitGroupEvent) TraceID() string {
	//TODO implement me
	panic("implement me")
}

func (c *WaitGroupEvent) TxHash() string {
	//TODO implement me
	panic("implement me")
}

func (c *WaitGroupEvent) ChainID() string {
	//TODO implement me
	panic("implement me")
}

func (c *WaitGroupEvent) Sequence() uint64 {
	//TODO implement me
	panic("implement me")
}

func (c WaitGroupEvent) SetSequenceID(s uint64) {
	//TODO implement me
	panic("implement me")
}

func (c WaitGroupEvent) BlockNr() int64 {
	//TODO implement me
	panic("implement me")
}

func (c WaitGroupEvent) StreamMessage() *eventspb.BusEvent {
	//TODO implement me
	panic("implement me")
}

// sequentialSqlStoreBroker : push events to each subscriber with a single go routine across all types
type sequentialSqlStoreBroker struct {
	typeToSubs          map[events.Type][]SqlBrokerSubscriber
	eventSource         eventSource
	chainInfo           ChainInfoI
	eventTypeBufferSize int
	log                 *logging.Logger
}

func newSequentialSqlStoreBroker(log *logging.Logger, chainInfo ChainInfoI,
	eventsource eventSource, eventTypeBufferSize int, subs []SqlBrokerSubscriber,
) *sequentialSqlStoreBroker {
	b := &sequentialSqlStoreBroker{
		typeToSubs:          map[events.Type][]SqlBrokerSubscriber{},
		eventSource:         eventsource,
		chainInfo:           chainInfo,
		eventTypeBufferSize: eventTypeBufferSize,
		log:                 log,
	}

	for _, s := range subs {
		for _, evtType := range s.Types() {
			b.typeToSubs[evtType] = append(b.typeToSubs[evtType], s)
		}
	}
	return b
}

func (b *sequentialSqlStoreBroker) Receive(ctx context.Context) error {
	if err := b.eventSource.Listen(); err != nil {
		return err
	}

	receiveCh, errCh := b.eventSource.Receive(ctx)

	start := time.Now()

	for e := range receiveCh {
		if err := checkChainID(b.chainInfo, e.ChainID()); err != nil {
			return err
		}

		if e.BlockNr() >= 2000 {
			now := time.Now()
			fmt.Printf("total time:%s", now.Sub(start))
		}

		fmt.Printf("Event:%s\n", e.StreamMessage())

		// If the event is a time event send it to all subscribers, this indicates a new block start (for now)
		if e.Type() == events.TimeUpdate {
			for _, subs := range b.typeToSubs {
				for _, sub := range subs {
					sub.Push(e)
				}
			}
		} else {
			if subs, ok := b.typeToSubs[e.Type()]; ok {
				for _, sub := range subs {
					sub.Push(e)
				}
			}
		}
	}

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
