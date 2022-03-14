package sqlstore

import (
	"context"
	"fmt"

	"code.vegaprotocol.io/data-node/entities"
	"github.com/georgysavva/scany/pgxscan"
)

type Rewards struct {
	*SQLStore
}

func NewRewards(sqlStore *SQLStore) *Rewards {
	r := &Rewards{
		SQLStore: sqlStore,
	}
	return r
}

func (rs *Rewards) Add(ctx context.Context, r entities.Reward) error {
	_, err := rs.pool.Exec(ctx,
		`INSERT INTO rewards(
			party_id,
			asset_id,
			epoch_id,
			amount,
			percent_of_total,
			vega_time)
		 VALUES ($1,  $2,  $3,  $4,  $5,  $6);`,
		r.PartyID, r.AssetID, r.EpochID, r.Amount, r.PercentOfTotal, r.VegaTime)
	return err
}

func (rs *Rewards) GetAll(ctx context.Context) ([]entities.Reward, error) {
	rewards := []entities.Reward{}
	err := pgxscan.Select(ctx, rs.pool, &rewards, `
		SELECT * from rewards;`)
	return rewards, err
}

func (rs *Rewards) Get(ctx context.Context,
	partyIDHex *string,
	assetIDHex *string,
	p *entities.Pagination,
) ([]entities.Reward, error) {
	query := `SELECT * from rewards`
	args := []interface{}{}
	if err := addRewardWhereClause(&query, &args, partyIDHex, assetIDHex); err != nil {
		return nil, err
	}

	if p != nil {
		query, args = paginateRewardQuery(query, args, *p)
	}

	rewards := []entities.Reward{}
	err := pgxscan.Select(ctx, rs.pool, &rewards, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying rewards: %w", err)
	}
	return rewards, nil
}

func (rs *Rewards) GetSummaries(ctx context.Context,
	partyIDHex *string, assetIDHex *string) ([]entities.RewardSummary, error) {

	query := `SELECT party_id, asset_id, sum(amount) as amount FROM rewards`
	args := []interface{}{}
	if err := addRewardWhereClause(&query, &args, partyIDHex, assetIDHex); err != nil {
		return nil, err
	}

	query = fmt.Sprintf("%s GROUP BY party_id, asset_id", query)

	summaries := []entities.RewardSummary{}
	err := pgxscan.Select(ctx, rs.pool, &summaries, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying rewards: %w", err)
	}
	return summaries, nil
}

// -------------------------------------------- Utility Methods

func addRewardWhereClause(queryPtr *string, args *[]interface{}, partyIDHex, assetIDHex *string) error {
	query := *queryPtr
	if partyIDHex != nil {
		partyID, err := entities.MakePartyID(*partyIDHex)
		if err != nil {
			return err
		}
		query = fmt.Sprintf("%s WHERE party_id=%s", query, nextBindVar(args, partyID))
	}

	if assetIDHex != nil {
		clause := "WHERE"
		if partyIDHex != nil {
			clause = "AND"
		}

		assetID := entities.MakeAssetID(*assetIDHex)
		query = fmt.Sprintf("%s %s asset_id=%s", query, clause, nextBindVar(args, assetID))
	}
	*queryPtr = query
	return nil
}

func paginateRewardQuery(query string, args []interface{}, p entities.Pagination) (string, []interface{}) {
	dir := "ASC"
	if p.Descending {
		dir = "DESC"
	}

	var limit interface{} = nil
	if p.Limit != 0 {
		limit = p.Limit
	}

	query = fmt.Sprintf(" %s ORDER BY epoch_id %s, party_id %s, asset_id %s LIMIT %s OFFSET %s",
		query, dir, dir, dir, nextBindVar(&args, limit), nextBindVar(&args, p.Skip))

	return query, args
}
