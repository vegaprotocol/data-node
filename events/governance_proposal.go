package events

import (
	"context"

	"code.vegaprotocol.io/data-node/proto"
	eventspb "code.vegaprotocol.io/data-node/proto/events/v1"
	"code.vegaprotocol.io/data-node/types"
)

type Proposal struct {
	*Base
	p types.Proposal
}

func NewProposalEvent(ctx context.Context, p types.Proposal) *Proposal {
	cpy := p.DeepClone()
	return &Proposal{
		Base: newBase(ctx, ProposalEvent),
		p:    *cpy,
	}
}

func (p *Proposal) Proposal() proto.Proposal {
	return *p.p.IntoProto()
}

// ProposalID - for combined subscriber, communal interface
func (p *Proposal) ProposalID() string {
	return p.p.Id
}

func (p Proposal) IsParty(id string) bool {
	return p.p.PartyId == id
}

// PartyID - for combined subscriber, communal interface
func (p *Proposal) PartyID() string {
	return p.p.PartyId
}

func (p Proposal) Proto() proto.Proposal {
	pr := p.p.IntoProto()
	return *pr
}

func (p Proposal) StreamMessage() *eventspb.BusEvent {
	return &eventspb.BusEvent{
		Id:    p.eventID(),
		Block: p.TraceID(),
		Type:  p.et.ToProto(),
		Event: &eventspb.BusEvent_Proposal{
			Proposal: p.p.IntoProto(),
		},
	}
}
