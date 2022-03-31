package entities_test

// No race condition checks on these tests, the channels are buffered to avoid actual issues
// we are aware that the tests themselves can be written in an unsafe way, but that's the tests
// not the code itsel. The behaviour of the tests is 100% reliable.
import (
	"context"
	"testing"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/vega/events"
	"code.vegaprotocol.io/vega/types/num"

	"github.com/stretchr/testify/assert"
)

func TestNewState(t *testing.T) {
	market := "market-id"
	party := "party1"

	position := entities.NewEmptyPosition(entities.NewMarketID(market), entities.NewPartyID(party))

	mp := marketPosition{party: party,
		size:   0,
		buy:    10,
		sell:   0,
		price:  num.Zero(),
		vwBuy:  num.Zero(),
		vwSell: num.Zero()}
	ctx := context.Background()
	ps := events.NewPositionStateEvent(ctx, mp, market)
	position.UpdateWithPositionState(ps)

	pp := position.ToProto()
	// This is an position with no values yet as nothing had traded
	assert.Equal(t, market, pp.MarketId)
	assert.Equal(t, party, pp.PartyId)
	assert.EqualValues(t, 0, pp.OpenVolume)
	assert.Equal(t, "0", pp.AverageEntryPrice)
	assert.Equal(t, "0", pp.RealisedPnl)
	assert.Equal(t, "0", pp.UnrealisedPnl)
	assert.EqualValues(t, 0, pp.UpdatedAt)

	mp.sell = 50
	ps2 := events.NewPositionStateEvent(ctx, mp, market)
	position.UpdateWithPositionState(ps2)

	pp = position.ToProto()
	// This is an position with no values yet as nothing had traded
	assert.Equal(t, market, pp.MarketId)
	assert.Equal(t, party, pp.PartyId)
	assert.EqualValues(t, 0, pp.OpenVolume)
	assert.Equal(t, "0", pp.AverageEntryPrice)
	assert.Equal(t, "0", pp.RealisedPnl)
	assert.Equal(t, "0", pp.UnrealisedPnl)
	assert.EqualValues(t, 0, pp.UpdatedAt)
}

func TestMultipleTradesOfSameSize(t *testing.T) {
	ctx := context.Background()
	market := "market-id"
	party := "party1"
	position := entities.NewEmptyPosition(entities.NewMarketID(market), entities.NewPartyID(party))
	ps := events.NewSettlePositionEvent(ctx, party, market, num.NewUint(1000), []events.TradeSettlement{
		tradeStub{
			size:  -1,
			price: num.NewUint(1000),
		},
		tradeStub{
			size:  -1,
			price: num.NewUint(1000),
		},
	}, 1, num.DecimalFromFloat(1))
	position.UpdateWithPositionSettlement(ps)
	pp := position.ToProto()
	// average entry price should be 1k
	assert.Equal(t, ps.Price().String(), pp.AverageEntryPrice)
}

func TestMultipleTradesAndLossSocializationPartyNoOpenVolume(t *testing.T) {
	ctx := context.Background()
	market := "market-id"
	party := "party1"
	position := entities.NewEmptyPosition(entities.NewMarketID(market), entities.NewPartyID(party))

	ps := events.NewSettlePositionEvent(ctx, party, market, num.NewUint(1000), []events.TradeSettlement{
		tradeStub{
			size:  2,
			price: num.NewUint(1000),
		},
		tradeStub{
			size:  -2,
			price: num.NewUint(1500),
		},
	}, 1, num.DecimalFromFloat(1))
	position.UpdateWithPositionSettlement(ps)
	pp := position.ToProto()
	assert.Equal(t, "1000", pp.RealisedPnl)

	// then we process the event for LossSocialization
	lsevt := events.NewLossSocializationEvent(ctx, party, market, num.NewUint(300), true, 1)
	position.UpdateWithLossSocialization(lsevt)
	pp = position.ToProto()
	assert.Equal(t, "700", pp.RealisedPnl)
	assert.Equal(t, "0", pp.UnrealisedPnl)
}

func TestDistressedPartyUpdate(t *testing.T) {
	ctx := context.Background()
	market := "market-id"
	party := "party1"
	position := entities.NewEmptyPosition(entities.NewMarketID(market), entities.NewPartyID(party))

	ps := events.NewSettlePositionEvent(ctx, party, market, num.NewUint(1000), []events.TradeSettlement{
		tradeStub{
			size:  2,
			price: num.NewUint(1000),
		},
		tradeStub{
			size:  3,
			price: num.NewUint(1200),
		},
	}, 1, num.DecimalFromFloat(1))
	position.UpdateWithPositionSettlement(ps)
	pp := position.ToProto()
	assert.Equal(t, "0", pp.RealisedPnl)
	assert.Equal(t, "-600", pp.UnrealisedPnl)

	// then we process the event for LossSocialization
	lsevt := events.NewLossSocializationEvent(ctx, party, market, num.NewUint(300), true, 1)
	position.UpdateWithLossSocialization(lsevt)
	pp = position.ToProto()
	assert.Equal(t, "-300", pp.RealisedPnl)
	assert.Equal(t, "-600", pp.UnrealisedPnl)

	// now assume this party is distressed, and we've taken all their funds
	sde := events.NewSettleDistressed(ctx, party, market, num.Zero(), num.NewUint(100), 1)
	position.UpdateWithSettleDestressed(sde)
	pp = position.ToProto()
	assert.Equal(t, "0", pp.UnrealisedPnl)
	assert.Equal(t, "-1000", pp.RealisedPnl)
}

func TestMultipleTradesAndLossSocializationPartyWithOpenVolume(t *testing.T) {
	ctx := context.Background()
	market := "market-id"
	party := "party1"
	position := entities.NewEmptyPosition(entities.NewMarketID(market), entities.NewPartyID(party))

	ps := events.NewSettlePositionEvent(ctx, party, market, num.NewUint(1000), []events.TradeSettlement{
		tradeStub{
			size:  2,
			price: num.NewUint(1000),
		},
		tradeStub{
			size:  3,
			price: num.NewUint(1200),
		},
	}, 1, num.DecimalFromFloat(1))
	position.UpdateWithPositionSettlement(ps)
	pp := position.ToProto()
	assert.Equal(t, "0", pp.RealisedPnl)
	assert.Equal(t, "-600", pp.UnrealisedPnl)

	// then we process the event for LossSocialization
	lsevt := events.NewLossSocializationEvent(ctx, party, market, num.NewUint(300), true, 1)
	position.UpdateWithLossSocialization(lsevt)
	pp = position.ToProto()
	assert.Equal(t, "-300", pp.RealisedPnl)
	assert.Equal(t, "-600", pp.UnrealisedPnl)
}

type tradeStub struct {
	size  int64
	price *num.Uint
}

func (t tradeStub) Size() int64 {
	return t.size
}

func (t tradeStub) Price() *num.Uint {
	return t.price.Clone()
}

type marketPosition struct {
	party  string
	size   int64
	buy    int64
	sell   int64
	price  *num.Uint
	vwBuy  *num.Uint
	vwSell *num.Uint
}

func (mp marketPosition) Buy() int64 {
	return mp.buy
}
func (mp marketPosition) Sell() int64 {
	return mp.sell
}
func (mp marketPosition) Party() string {
	return mp.party
}
func (mp marketPosition) Size() int64 {
	return mp.size
}
func (mp marketPosition) Price() *num.Uint {
	return mp.price
}
func (mp marketPosition) VWBuy() *num.Uint {
	return mp.vwBuy
}
func (mp marketPosition) VWSell() *num.Uint {
	return mp.vwSell
}
