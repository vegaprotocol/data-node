package subscribers

import (
	"context"
	"sync"

	"code.vegaprotocol.io/data-node/events"
	"code.vegaprotocol.io/data-node/logging"
	types "code.vegaprotocol.io/data-node/proto/vega"
)

type GovernanceDataSub struct {
	*Base
	mu        sync.RWMutex
	proposals map[string]types.Proposal
	byPID     map[string]*types.GovernanceData
	all       []*types.GovernanceData
	log       *logging.Logger
}

func NewGovernanceDataSub(ctx context.Context, log *logging.Logger, ack bool) *GovernanceDataSub {
	gd := &GovernanceDataSub{
		Base:      NewBase(ctx, 10, ack),
		proposals: map[string]types.Proposal{},
		byPID:     map[string]*types.GovernanceData{},
		all:       []*types.GovernanceData{},
		log:       log,
	}
	if gd.isRunning() {
		go gd.loop(gd.ctx)
	}
	return gd
}

func (g *GovernanceDataSub) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			g.Halt()
			return
		case e := <-g.ch:
			if g.isRunning() {
				g.Push(e...)
			}
		}
	}
}

func (g *GovernanceDataSub) Push(evts ...events.Event) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, e := range evts {
		switch et := e.(type) {
		case PropE:
			prop := et.Proposal()
			gd := g.getData(prop.Id)
			g.proposals[prop.Id] = prop
			gd.Proposal = &prop
		case VoteE:
			vote := et.Vote()
			gd := g.getData(vote.ProposalId)
			if vote.Value == types.Vote_VALUE_YES {
				gd.Yes = append(gd.Yes, &vote)
				delete(gd.NoParty, vote.PartyId)
				gd.YesParty[vote.PartyId] = &vote
			} else {
				gd.No = append(gd.No, &vote)
				delete(gd.YesParty, vote.PartyId)
				gd.NoParty[vote.PartyId] = &vote
			}
		default:
			g.log.Panic("Unknown event type in governance subscriber", logging.String("Type", et.Type().String()))
		}
	}
}

// Filter - get filtered proposals, value receiver so no data races
// uniqueVotes will replace the slice (containing older/duplicate votes) with only the latest votes
// for each participant
func (g *GovernanceDataSub) Filter(uniqueVotes bool, params ...ProposalFilter) []*types.GovernanceData {
	g.mu.RLock()
	defer g.mu.RUnlock()
	ret := []*types.GovernanceData{}
	for id, p := range g.proposals {
		add := true
		for _, f := range params {
			if !f(p) {
				add = false
				break
			}
		}
		if add {
			// create a copy
			gd := *g.byPID[id]
			if uniqueVotes {
				gd.Yes = make([]*types.Vote, 0, len(gd.YesParty))
				for _, v := range gd.YesParty {
					gd.Yes = append(gd.Yes, v)
				}
				gd.No = make([]*types.Vote, 0, len(gd.NoParty))
				for _, v := range gd.NoParty {
					gd.No = append(gd.No, v)
				}
			}
			//  add to the return value
			ret = append(ret, &gd)
		}
	}
	return ret
}

func (g *GovernanceDataSub) getData(id string) *types.GovernanceData {
	if gd, ok := g.byPID[id]; ok {
		return gd
	}
	gd := &types.GovernanceData{
		Yes:      []*types.Vote{},
		No:       []*types.Vote{},
		YesParty: map[string]*types.Vote{},
		NoParty:  map[string]*types.Vote{},
	}
	g.byPID[id] = gd
	g.all = append(g.all, gd)
	return gd
}

func (g *GovernanceDataSub) Types() []events.Type {
	return []events.Type{
		events.ProposalEvent,
		events.VoteEvent,
	}
}
