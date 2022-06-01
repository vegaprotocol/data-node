package sqlstore

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/metrics"
	"github.com/georgysavva/scany/pgxscan"
)

var ErrPositionNotFound = errors.New("party not found")

type Positions struct {
	*ConnectionSource
	cache     map[entities.MarketID]map[entities.PartyID]entities.Position
	cacheLock sync.Mutex
	batcher   MapBatcher[entities.PositionKey, entities.Position]
}

func NewPositions(connectionSource *ConnectionSource) *Positions {
	a := &Positions{
		ConnectionSource: connectionSource,
		cache:            map[entities.MarketID]map[entities.PartyID]entities.Position{},
		cacheLock:        sync.Mutex{},
		batcher: NewMapBatcher[entities.PositionKey, entities.Position](
			"positions",
			entities.PositionColumns),
	}
	return a
}

func (ps *Positions) Flush(ctx context.Context) error {
	defer metrics.StartSQLQuery("Positions", "FlushTest")()
	return ps.batcher.Flush(ctx, ps.pool)
}

func (ps *Positions) Add(ctx context.Context, p entities.Position) error {
	ps.cacheLock.Lock()
	defer ps.cacheLock.Unlock()
	ps.batcher.Add(p)
	ps.updateCache(p)
	return nil
}

func (ps *Positions) GetByMarketAndParty(ctx context.Context,
	marketID entities.MarketID,
	partyID entities.PartyID,
) (entities.Position, error) {
	ps.cacheLock.Lock()
	defer ps.cacheLock.Unlock()

	position, found := ps.checkCache(marketID, partyID)
	if found {
		return position, nil
	}

	defer metrics.StartSQLQuery("Positions", "GetByMarketAndParty")()
	err := pgxscan.Get(ctx, ps.Connection, &position,
		`SELECT * FROM positions_current WHERE market_id=$1 AND party_id=$2`,
		marketID, partyID)

	if err == nil {
		ps.updateCache(position)
	}

	if pgxscan.NotFound(err) {
		return position, fmt.Errorf("'%v/%v': %w", marketID, partyID, ErrPositionNotFound)
	}

	return position, err
}

func (ps *Positions) GetByMarket(ctx context.Context, marketID entities.MarketID, p *entities.OffsetPagination) ([]entities.Position, error) {
	defer metrics.StartSQLQuery("Positions", "GetByMarket")()
	query := `SELECT * FROM positions_current WHERE market_id=$1`
	args := []interface{}{marketID}
	if p != nil {
		query, args = paginatePositionQuery(query, args, *p)
	}
	positions := []entities.Position{}
	err := pgxscan.Select(ctx, ps.Connection, &positions,
		query,
		args)
	return positions, err
}

func (ps *Positions) GetByParty(ctx context.Context, partyID entities.PartyID, p *entities.OffsetPagination) ([]entities.Position, error) {
	defer metrics.StartSQLQuery("Positions", "GetByParty")()
	query := `SELECT * FROM positions_current WHERE party_id=$1`
	args := []interface{}{partyID}
	if p != nil {
		query, args = paginatePositionQuery(query, args, *p)
	}
	positions := []entities.Position{}
	err := pgxscan.Select(ctx, ps.Connection, &positions,
		query,
		args)
	return positions, err
}

func (ps *Positions) GetAll(ctx context.Context) ([]entities.Position, error) {
	defer metrics.StartSQLQuery("Positions", "GetAll")()

	positions := []entities.Position{}
	err := pgxscan.Select(ctx, ps.Connection, &positions,
		`SELECT * FROM positions_current`)
	return positions, err
}

func paginatePositionQuery(query string, args []interface{}, p entities.OffsetPagination) (string, []interface{}) {
	dir := "ASC"
	if p.Descending {
		dir = "DESC"
	}

	var limit interface{} = nil
	if p.Limit != 0 {
		limit = p.Limit
	}

	query = fmt.Sprintf(" %s ORDER BY vega_time %s, id %s LIMIT %s OFFSET %s",
		query, dir, dir, nextBindVar(&args, limit), nextBindVar(&args, p.Skip))

	return query, args
}

func (ps *Positions) updateCache(p entities.Position) {
	if _, ok := ps.cache[p.MarketID]; !ok {
		ps.cache[p.MarketID] = map[entities.PartyID]entities.Position{}
	}

	ps.cache[p.MarketID][p.PartyID] = p
}

func (ps *Positions) checkCache(marketID entities.MarketID, partyID entities.PartyID) (entities.Position, bool) {
	if _, ok := ps.cache[marketID]; !ok {
		return entities.Position{}, false
	}

	pos, ok := ps.cache[marketID][partyID]
	if !ok {
		return entities.Position{}, false
	}
	return pos, true
}
