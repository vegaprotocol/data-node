package sqlstore

import (
	"context"
	"fmt"
	"strings"

	"code.vegaprotocol.io/data-node/entities"
	"github.com/georgysavva/scany/pgxscan"
)

type Delegations struct {
	*SQLStore
}

func NewDelegations(sqlStore *SQLStore) *Delegations {
	d := &Delegations{
		SQLStore: sqlStore,
	}
	return d
}

func (ds *Delegations) Add(ctx context.Context, d entities.Delegation) error {
	_, err := ds.pool.Exec(ctx,
		`INSERT INTO delegations(
			party_id,
			node_id,
			epoch_id,
			amount,
			vega_time)
		 VALUES ($1,  $2,  $3,  $4,  $5);`,
		d.PartyID, d.NodeID, d.EpochID, d.Amount, d.VegaTime)
	return err
}

func (ds *Delegations) GetAll(ctx context.Context) ([]entities.Delegation, error) {
	delegations := []entities.Delegation{}
	err := pgxscan.Select(ctx, ds.pool, &delegations, `
		SELECT * from delegations;`)
	return delegations, err
}

func (ds *Delegations) Get(ctx context.Context,
	partyIDHex *string,
	nodeIDHex *string,
	epochID *int64,
	p *entities.Pagination,
) ([]entities.Delegation, error) {
	query := `SELECT * from delegations`
	args := []interface{}{}

	conditions := []string{}

	if partyIDHex != nil {
		partyID, err := entities.MakePartyID(*partyIDHex)
		if err != nil {
			return []entities.Delegation{}, err
		}
		conditions = append(conditions, fmt.Sprintf("party_id=%s", nextBindVar(&args, partyID)))
	}

	if nodeIDHex != nil {
		nodeID, err := entities.MakeNodeID(*nodeIDHex)
		if err != nil {
			return []entities.Delegation{}, err
		}
		conditions = append(conditions, fmt.Sprintf("node_id=%s", nextBindVar(&args, nodeID)))
	}

	if epochID != nil {
		conditions = append(conditions, fmt.Sprintf("epoch_id=%s", nextBindVar(&args, *epochID)))
	}

	if len(conditions) > 0 {
		query = fmt.Sprintf("%s WHERE %s", query, strings.Join(conditions, " AND "))
	}

	if p != nil {
		query, args = paginateDelegationQuery(query, args, *p)
	}

	delegations := []entities.Delegation{}
	err := pgxscan.Select(ctx, ds.pool, &delegations, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying delegations: %w", err)
	}
	return delegations, nil
}

// -------------------------------------------- Utility Methods

func paginateDelegationQuery(query string, args []interface{}, p entities.Pagination) (string, []interface{}) {
	dir := "ASC"
	if p.Descending {
		dir = "DESC"
	}

	var limit interface{} = nil
	if p.Limit != 0 {
		limit = p.Limit
	}

	query = fmt.Sprintf(" %s ORDER BY epoch_id %s, party_id %s, node_id %s LIMIT %s OFFSET %s",
		query, dir, dir, dir, nextBindVar(&args, limit), nextBindVar(&args, p.Skip))

	return query, args
}
