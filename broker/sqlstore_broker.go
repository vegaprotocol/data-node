package broker

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/data-node/metrics"
	"code.vegaprotocol.io/data-node/sqlstore"
	"code.vegaprotocol.io/data-node/sqlsubscribers"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vega/events"

	"math"
)

type SqlBrokerSubscriber interface {
	Push(val events.Event) error
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
	//if config.UseSequentialSqlStoreBroker {
	return newSequentialSqlStoreBroker(log, chainInfo, eventsource, eventTypeBufferSize, subs, bool(config.PanicOnError))
	//	} else {
	//return newConcurrentSqlStoreBroker(log, chainInfo, eventsource, eventTypeBufferSize, subs, bool(config.PanicOnError))
	//	}
}

// concurrentSqlStoreBroker : push events to each subscriber with a go-routine per type
type concurrentSqlStoreBroker struct {
	log                 *logging.Logger
	typeToSubs          map[events.Type][]SqlBrokerSubscriber
	typeToEvtCh         map[events.Type]chan events.Event
	eventSource         eventSource
	chainInfo           ChainInfoI
	eventTypeBufferSize int
	panicOnError        bool
}

func newConcurrentSqlStoreBroker(log *logging.Logger, chainInfo ChainInfoI, eventsource eventSource, eventTypeBufferSize int,
	subs []SqlBrokerSubscriber, panicOnError bool,
) *concurrentSqlStoreBroker {
	b := &concurrentSqlStoreBroker{
		log:                 log,
		typeToSubs:          map[events.Type][]SqlBrokerSubscriber{},
		typeToEvtCh:         map[events.Type]chan events.Event{},
		eventSource:         eventsource,
		chainInfo:           chainInfo,
		eventTypeBufferSize: 10000,
		panicOnError:        panicOnError,
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

	start := time.Now()
	blockStart := time.Now()
	eventCount := 0
	typeToCount := map[events.Type]int{}

	for e := range receiveCh {
		if err := checkChainID(b.chainInfo, e.ChainID()); err != nil {
			return err
		}

		eventCount++
		typeToCount[e.Type()] = typeToCount[e.Type()] + 1
		if e.BlockNr() >= 60000 {
			now := time.Now()
			taken := now.Sub(start)
			fmt.Printf("total time:%s", taken)
		}

		// If the event is a time event send it to all type channels, this indicates a new block start (for now)
		if e.Type() == events.TimeUpdate {
			timeUpdate := e.(sqlsubscribers.TimeUpdateEvent)

			now := time.Now()
			typeCount := ""
			for t, c := range typeToCount {
				typeCount += fmt.Sprintf(",EVENT%s=%d", t.String(), c)
			}

			fmt.Printf("Time=%s,Block=%d,VegaTime=%s,BlockProcessingTime:%s,BlockEventCount=%d%s\n", now, e.BlockNr(),
				timeUpdate.Time(), now.Sub(blockStart), eventCount, typeCount)
			blockStart = now
			eventCount = 0
			typeToCount = map[events.Type]int{}

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
							err := sub.Push(time)
							if err != nil {
								b.OnPushEventError(time, err)
							}
						}
					} else {
						for _, sub := range subs {
							select {
							case <-ctx.Done():
								return
							default:
								err := sub.Push(evt)
								if err != nil {
									b.OnPushEventError(evt, err)
								}
							}
						}
					}
				}
			}
		}(ch, b.typeToSubs[t])
	}
}

func (b *concurrentSqlStoreBroker) OnPushEventError(evt events.Event, err error) {
	errMsg := fmt.Sprintf("failed to process event %v error:%+v", evt.StreamMessage(), err)
	//	if b.panicOnError {
	//		b.log.Panic(errMsg)
	//	} else {
	b.log.Error(errMsg)
	//	}

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
	log                 *logging.Logger
	typeToSubs          map[events.Type][]SqlBrokerSubscriber
	eventSource         eventSource
	chainInfo           ChainInfoI
	eventTypeBufferSize int
	panicOnError        bool
}

func newSequentialSqlStoreBroker(log *logging.Logger, chainInfo ChainInfoI,
	eventsource eventSource, eventTypeBufferSize int, subs []SqlBrokerSubscriber,
	panicOnError bool,
) *sequentialSqlStoreBroker {
	b := &sequentialSqlStoreBroker{
		log:                 log,
		typeToSubs:          map[events.Type][]SqlBrokerSubscriber{},
		eventSource:         eventsource,
		chainInfo:           chainInfo,
		eventTypeBufferSize: eventTypeBufferSize,
		panicOnError:        panicOnError,
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
	blockStart := time.Now()
	eventCount := 0
	typeToCount := map[events.Type]int{}

	for e := range receiveCh {
		if err := checkChainID(b.chainInfo, e.ChainID()); err != nil {
			return err
		}

		eventCount++
		typeToCount[e.Type()] = typeToCount[e.Type()] + 1

		if e.BlockNr() >= 60000 {
			now := time.Now()
			taken := now.Sub(start)
			fmt.Printf("total time:%s", taken)
		}

		if sqlstore.GlobalTx == nil {
			var err error
			sqlstore.GlobalTx, err = sqlstore.Pool.Begin(context.Background())
			if err != nil {
				panic(err)
			}
		}

		metrics.EventCounterInc(e.Type().String())
		// If the event is a time event send it to all subscribers, this indicates a new block start (for now)
		if e.Type() == events.TimeUpdate {

			for _, subs := range b.typeToSubs {
				for _, sub := range subs {
					b.push(sub, e)
				}
			}

			err := sqlstore.GlobalTx.Commit(context.Background())
			if err != nil {
				b.log.Errorf("commitly error:%v\n", err)
				sqlstore.GlobalTx.Rollback(context.TODO())
			} else {
				b.log.Info("commited transaction succesfully")
			}

			sqlstore.GlobalTx = nil

			timeUpdate := e.(sqlsubscribers.TimeUpdateEvent)

			now := time.Now()
			typeCount := ""
			for t, c := range typeToCount {
				typeCount += fmt.Sprintf(",EVENT%s=%d", t.String(), c)
			}

			fmt.Printf("Time=%s,Block=%d,VegaTime=%s,BlockProcessingTime:%s,BlockEventCount=%d%s\n", now, e.BlockNr(),
				timeUpdate.Time(), now.Sub(blockStart), eventCount, typeCount)
			blockStart = now
			eventCount = 0
			typeToCount = map[events.Type]int{}

			metrics.BlockCounterInc()

		} else {
			if subs, ok := b.typeToSubs[e.Type()]; ok {
				for _, sub := range subs {
					b.push(sub, e)
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

func (b *sequentialSqlStoreBroker) push(sub SqlBrokerSubscriber, e events.Event) {
	sub_name := reflect.TypeOf(sub).Elem().Name()
	timer := metrics.NewTimeCounter("sql", sub_name, e.Type().String())
	err := sub.Push(e)
	if err != nil {
		b.OnPushEventError(e, err)
	}
	timer.EventTimeCounterAdd()
}

func (b *sequentialSqlStoreBroker) OnPushEventError(evt events.Event, err error) {
	errMsg := fmt.Sprintf("failed to process event %v error:%+v", evt.StreamMessage(), err)
	//	if b.panicOnError {
	//		b.log.Panic(errMsg)
	//	} else {
	b.log.Error(errMsg)
	//	}

}
