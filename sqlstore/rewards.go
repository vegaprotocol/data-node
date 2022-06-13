package sqlstore

import (
	"context"
	"fmt"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/metrics"
	"github.com/georgysavva/scany/pgxscan"
)

type Rewards struct {
	*ConnectionSource
}

func NewRewards(connectionSource *ConnectionSource) *Rewards {
	r := &Rewards{
		ConnectionSource: connectionSource,
	}
	return r
}

func (rs *Rewards) Add(ctx context.Context, r entities.Reward) error {
	defer metrics.StartSQLQuery("Rewards", "Add")()
	_, err := rs.Connection.Exec(ctx,
		`INSERT INTO rewards(
			party_id,
			asset_id,
			market_id,
			reward_type,
			epoch_id,
			amount,
			percent_of_total,
			timestamp,
			vega_time)
		 VALUES ($1,  $2,  $3,  $4,  $5,  $6, $7, $8, $9);`,
		r.PartyID, r.AssetID, r.MarketID, r.RewardType, r.EpochID, r.Amount, r.PercentOfTotal, r.Timestamp, r.VegaTime)
	return err
}

func (rs *Rewards) GetAll(ctx context.Context) ([]entities.Reward, error) {
	defer metrics.StartSQLQuery("Rewards", "GetAll")()
	rewards := []entities.Reward{}
	err := pgxscan.Select(ctx, rs.Connection, &rewards, `
		SELECT * from rewards;`)
	return rewards, err
}

func (rs *Rewards) Get(ctx context.Context,
	partyIDHex *string,
	assetIDHex *string,
	pagination entities.Pagination,
) ([]entities.Reward, entities.PageInfo, error) {
	switch p := pagination.(type) {
	case entities.OffsetPagination:
		return rs.getByOffset(ctx, partyIDHex, assetIDHex, &p)
	case *entities.OffsetPagination:
		return rs.getByOffset(ctx, partyIDHex, assetIDHex, p)
	case entities.CursorPagination:
		return rs.getByCursor(ctx, partyIDHex, assetIDHex, p)
	case *entities.CursorPagination:
		return rs.getByCursor(ctx, partyIDHex, assetIDHex, *p)
	case nil:
		return rs.getByOffset(ctx, partyIDHex, assetIDHex, nil)
	default:
		return nil, entities.PageInfo{}, fmt.Errorf("unsupported pagination type: %T", p)
	}
}

func (rs *Rewards) getByOffset(ctx context.Context, partyIDHex, assetIDHex *string, p *entities.OffsetPagination) ([]entities.Reward, entities.PageInfo, error) {
	query, args, err := selectRewards(partyIDHex, assetIDHex)
	if err != nil {
		return nil, entities.PageInfo{}, err
	}

	if p != nil {
		order_cols := []string{"epoch_id", "party_id", "asset_id"}
		query, args = orderAndPaginateQuery(query, order_cols, *p, args...)
	}

	rewards := []entities.Reward{}
	defer metrics.StartSQLQuery("Rewards", "Get")()
	err = pgxscan.Select(ctx, rs.Connection, &rewards, query, args...)
	if err != nil {
		return nil, entities.PageInfo{}, fmt.Errorf("querying rewards: %w", err)
	}
	return rewards, entities.PageInfo{}, nil

}

func (rs *Rewards) getByCursor(ctx context.Context, partyIDHex, assetIDHex *string, p entities.CursorPagination) ([]entities.Reward, entities.PageInfo, error) {
	query, args, err := selectRewards(partyIDHex, assetIDHex)
	if err != nil {
		return nil, entities.PageInfo{}, err
	}

	sorting, cmp, cursor := extractPaginationInfo(p)
	cursorParams := []CursorQueryParameter{
		NewCursorQueryParameter("epoch_id", sorting, cmp, cursor),
	}

	query, args = orderAndPaginateWithCursor(query, p, cursorParams, args...)

	rewards := []entities.Reward{}
	if err := pgxscan.Select(ctx, rs.Connection, &rewards, query, args...); err != nil {
		return nil, entities.PageInfo{}, fmt.Errorf("querying rewards: %w", err)
	}

	pagedData, pageInfo := entities.PageEntities(rewards, p)
	return pagedData, pageInfo, nil
}

func selectRewards(partyIDHex, assetIDHex *string) (string, []interface{}, error) {
	query := `SELECT * from rewards`
	args := []interface{}{}
	if err := addRewardWhereClause(&query, &args, partyIDHex, assetIDHex); err != nil {
		return "", nil, err
	}

	return query, args, nil
}

func (rs *Rewards) GetSummaries(ctx context.Context,
	partyIDHex *string, assetIDHex *string,
) ([]entities.RewardSummary, error) {
	query := `SELECT party_id, asset_id, sum(amount) as amount FROM rewards`
	args := []interface{}{}
	if err := addRewardWhereClause(&query, &args, partyIDHex, assetIDHex); err != nil {
		return nil, err
	}

	query = fmt.Sprintf("%s GROUP BY party_id, asset_id", query)

	summaries := []entities.RewardSummary{}
	defer metrics.StartSQLQuery("Rewards", "GetSummaries")()
	err := pgxscan.Select(ctx, rs.Connection, &summaries, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying rewards: %w", err)
	}
	return summaries, nil
}

// -------------------------------------------- Utility Methods

func addRewardWhereClause(queryPtr *string, args *[]interface{}, partyIDHex, assetIDHex *string) error {
	query := *queryPtr
	if partyIDHex != nil {
		partyID := entities.NewPartyID(*partyIDHex)
		query = fmt.Sprintf("%s WHERE party_id=%s", query, nextBindVar(args, partyID))
	}

	if assetIDHex != nil {
		clause := "WHERE"
		if partyIDHex != nil {
			clause = "AND"
		}

		assetID := entities.ID(*assetIDHex)
		query = fmt.Sprintf("%s %s asset_id=%s", query, clause, nextBindVar(args, assetID))
	}
	*queryPtr = query
	return nil
}
