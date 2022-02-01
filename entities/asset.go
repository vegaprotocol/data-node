package entities

import (
	"encoding/hex"
	"time"

	"github.com/shopspring/decimal"
)

type Asset struct {
	ID            []byte
	Name          string
	Symbol        string
	TotalSupply   decimal.Decimal // Maybe num.Uint if we can figure out how to add support to pgx
	Decimals      int
	Quantum       int
	Source        string
	ERC20Contract string
	VegaTime      time.Time
}

func (a Asset) HexId() string {
	return hex.EncodeToString(a.ID)
}
