package sqlstore

import (
	"context"
	"fmt"

	"code.vegaprotocol.io/data-node/entities"
	"github.com/georgysavva/scany/pgxscan"
)

type RiskFactors struct {
	*ConnectionSource
}

const (
	sqlRiskFactorColumns = `market_id, short, long, vega_time`
)

func NewRiskFactors(sqlStore *ConnectionSource) *RiskFactors {
	return &RiskFactors{
		ConnectionSource: sqlStore,
	}
}

func (rf *RiskFactors) Upsert(ctx context.Context, factor *entities.RiskFactor) error {
	query := fmt.Sprintf(`insert into risk_factors (%s)
values ($1, $2, $3, $4)
on conflict (market_id, vega_time) do update
set 
	short=EXCLUDED.short,
	long=EXCLUDED.long`, sqlRiskFactorColumns)

	if _, err := rf.Connection.Exec(ctx, query, factor.MarketID, factor.Short, factor.Long, factor.VegaTime); err != nil {
		err = fmt.Errorf("could not insert risk factor into database: %w", err)
		return err
	}

	return nil
}

func (rf *RiskFactors) GetMarketRiskFactors(ctx context.Context, marketID string) (entities.RiskFactor, error) {
	var riskFactor entities.RiskFactor
	var bindVars []interface{}

	query := fmt.Sprintf(`select %s
		from risk_factors
		where market_id = %s`, sqlRiskFactorColumns, nextBindVar(&bindVars, entities.NewMarketID(marketID)))

	err := pgxscan.Get(ctx, rf.Connection, &riskFactor, query, bindVars...)

	return riskFactor, err
}
