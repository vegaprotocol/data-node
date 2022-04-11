package entities

import (
	"time"

	"code.vegaprotocol.io/protos/vega"
	"code.vegaprotocol.io/vega/events"
	"code.vegaprotocol.io/vega/types/num"
	"github.com/shopspring/decimal"
)

type positionSettlement interface {
	Price() *num.Uint
	PositionFactor() num.Decimal
	Trades() []events.TradeSettlement
}

type lossSocialization interface {
	Amount() *num.Int
}

type settleDestressed interface {
	Margin() *num.Uint
}

type updatePosition interface {
	Size() int64
}

type Position struct {
	MarketID          MarketID
	PartyID           PartyID
	OpenVolume        int64
	RealisedPnl       decimal.Decimal
	UnrealisedPnl     decimal.Decimal
	AverageEntryPrice decimal.Decimal
	Loss              decimal.Decimal // what the party lost because of loss socialization
	Adjustment        decimal.Decimal // what a party was missing which triggered loss socialization
	VegaTime          time.Time
}

func NewEmptyPosition(marketID MarketID, partyID PartyID) Position {
	return Position{
		MarketID:          marketID,
		PartyID:           partyID,
		OpenVolume:        0,
		RealisedPnl:       decimal.Zero,
		UnrealisedPnl:     decimal.Zero,
		AverageEntryPrice: decimal.Zero,
		Loss:              decimal.Zero,
		Adjustment:        decimal.Zero,
	}
}

func (p *Position) UpdateWithPositionSettlement(e positionSettlement) {
	for _, t := range e.Trades() {
		openedVolume, closedVolume := calculateOpenClosedVolume(p.OpenVolume, t.Size())
		// Deal with any volume we have closed
		realisedPnlDelta := num.DecimalFromUint(t.Price()).Sub(p.AverageEntryPrice).Mul(num.DecimalFromInt64(closedVolume)).Div(e.PositionFactor())
		p.RealisedPnl = p.RealisedPnl.Add(realisedPnlDelta)
		p.OpenVolume -= closedVolume

		// Then with any we have opened
		p.AverageEntryPrice = updateVWAP(p.AverageEntryPrice, p.OpenVolume, openedVolume, t.Price())
		p.OpenVolume += openedVolume
	}
	p.mtm(e.Price(), e.PositionFactor())
}

func (p *Position) UpdateWithLossSocialization(e lossSocialization) {
	amountLoss := num.DecimalFromInt(e.Amount())

	if amountLoss.IsNegative() {
		p.Loss = p.Loss.Add(amountLoss)
	} else {
		p.Adjustment = p.Adjustment.Add(amountLoss)
	}

	p.RealisedPnl = p.RealisedPnl.Add(amountLoss)
}

func (p *Position) UpdateWithSettleDestressed(e settleDestressed) {
	margin := num.DecimalFromUint(e.Margin())
	p.RealisedPnl = p.RealisedPnl.Add(p.UnrealisedPnl)
	p.RealisedPnl = p.RealisedPnl.Sub(margin) // realised P&L includes whatever we had in margin account at this point
	p.UnrealisedPnl = decimal.Zero
	p.AverageEntryPrice = decimal.Zero // @TODO average entry price shouldn't be affected(?)
	p.AverageEntryPrice = decimal.Zero
	p.OpenVolume = 0
}

func (p *Position) UpdateWithPositionState(e updatePosition) {
	p.OpenVolume = e.Size()
}

func (p *Position) ToProto() *vega.Position {
	var timestamp int64
	if !p.VegaTime.IsZero() {
		timestamp = p.VegaTime.UnixNano()
	}
	return &vega.Position{
		MarketId:          p.MarketID.String(),
		PartyId:           p.PartyID.String(),
		OpenVolume:        p.OpenVolume,
		RealisedPnl:       p.RealisedPnl.Round(0).String(),
		UnrealisedPnl:     p.UnrealisedPnl.Round(0).String(),
		AverageEntryPrice: p.AverageEntryPrice.Round(0).String(),
		UpdatedAt:         timestamp,
	}
}

func (p *Position) AverageEntryPriceUint() *num.Uint {
	uint, overflow := num.UintFromDecimal(p.AverageEntryPrice)
	if overflow {
		panic("couldn't convert average entry price from decimal to uint")
	}
	return uint
}

func (p *Position) mtm(markPrice *num.Uint, positionFactor num.Decimal) {
	if p.OpenVolume == 0 {
		p.UnrealisedPnl = num.DecimalZero()
		return
	}
	markPriceDec := num.DecimalFromUint(markPrice)
	openVolumeDec := num.DecimalFromInt64(p.OpenVolume)

	p.UnrealisedPnl = openVolumeDec.Mul(markPriceDec.Sub(p.AverageEntryPrice)).Div(positionFactor)
}

func calculateOpenClosedVolume(currentOpenVolume, tradedVolume int64) (int64, int64) {
	if currentOpenVolume != 0 && ((currentOpenVolume > 0) != (tradedVolume > 0)) {
		var closedVolume int64
		if absUint64(tradedVolume) > absUint64(currentOpenVolume) {
			closedVolume = currentOpenVolume
		} else {
			closedVolume = -tradedVolume
		}
		return tradedVolume + closedVolume, closedVolume
	}
	return tradedVolume, 0
}

func absUint64(v int64) uint64 {
	if v < 0 {
		v *= -1
	}
	return uint64(v)
}

func updateVWAP(vwap num.Decimal, volume int64, addVolume int64, addPrice *num.Uint) num.Decimal {
	if volume+addVolume == 0 {
		return num.DecimalZero()
	}

	volumeDec := num.DecimalFromInt64(volume)
	addVolumeDec := num.DecimalFromInt64(addVolume)
	addPriceDec := num.DecimalFromUint(addPrice)

	return vwap.Mul(volumeDec).Add(addPriceDec.Mul(addVolumeDec)).Div(volumeDec.Add(addVolumeDec))
}