package events

import (
	"context"

	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/data-node/types/num"
)

type LossSoc struct {
	*Base
	partyID  string
	marketID string
	amount   int64
	ts       int64
}

func NewLossSocializationEvent(ctx context.Context, partyID, marketID string, amount int64, ts int64) *LossSoc {
	return &LossSoc{
		Base:     newBase(ctx, LossSocializationEvent),
		partyID:  partyID,
		marketID: marketID,
		amount:   amount,
		ts:       ts,
	}
}

func (l LossSoc) IsParty(id string) bool {
	return l.partyID == id
}

func (l LossSoc) PartyID() string {
	return l.partyID
}

func (l LossSoc) MarketID() string {
	return l.marketID
}

func (l LossSoc) Amount() int64 {
	return l.amount
}

func (l LossSoc) AmountLost() int64 {
	return l.amount
}

func (l LossSoc) Timestamp() int64 {
	return l.ts
}

func (l LossSoc) Proto() eventspb.LossSocialization {
	return eventspb.LossSocialization{
		MarketId: l.marketID,
		PartyId:  l.partyID,
		Amount:   l.amount,
	}
}

func (l LossSoc) StreamMessage() *eventspb.BusEvent {
	p := l.Proto()
	return &eventspb.BusEvent{
		Id:    l.eventID(),
		Block: l.TraceID(),
		Type:  l.et.ToProto(),
		Event: &eventspb.BusEvent_LossSocialization{
			LossSocialization: &p,
		},
	}
}

func LossSocializationEventFromStream(ctx context.Context, be *eventspb.BusEvent) *LossSoc {
	return &LossSoc{
		Base:     newBaseFromStream(ctx, LossSocializationEvent, be),
		partyID:  be.GetLossSocialization().PartyId,
		marketID: be.GetLossSocialization().MarketId,
		amount:   be.GetLossSocialization().Amount,
	}
}
