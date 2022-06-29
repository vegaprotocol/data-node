// Copyright (c) 2022 Gobalsky Labs Limited
//
// Use of this software is governed by the Business Source License included
// in the LICENSE file and at https://www.mariadb.com/bsl11.
//
// Change Date: 18 months from the later of the date of the first publicly
// available Distribution of this version of the repository, and 25 June 2022.
//
// On the date above, in accordance with the Business Source License, use
// of this software will be governed by version 3 or later of the GNU General
// Public License.

package sqlstore

import (
	"context"
	"fmt"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/metrics"
	v2 "code.vegaprotocol.io/protos/data-node/api/v2"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
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

func (rs *Rewards) GetByCursor(ctx context.Context,
	partyIDHex *string,
	assetIDHex *string,
	pagination entities.CursorPagination,
) entities.ConnectionData[*v2.RewardEdge, entities.Reward] {
	var connectionData entities.ConnectionData[*v2.RewardEdge, entities.Reward]
	selectQuery, countQuery, args := selectRewards(partyIDHex, assetIDHex)

	batch := pgx.Batch{}
	batch.Queue(countQuery, args...)

	sorting, cmp, cursor := extractPaginationInfo(pagination)
	rc := &entities.RewardCursor{}
	if cursor != "" {
		err := rc.Parse(cursor)
		if err != nil {
			connectionData.Err = fmt.Errorf("parsing cursor: %w", err)
			return connectionData
		}
	}
	cursorParams := []CursorQueryParameter{
		NewCursorQueryParameter("party_id", sorting, cmp, entities.NewPartyID(rc.PartyID)),
		NewCursorQueryParameter("asset_id", sorting, cmp, entities.NewAssetID(rc.AssetID)),
		NewCursorQueryParameter("epoch_id", sorting, cmp, rc.EpochID),
	}

	selectQuery, args = orderAndPaginateWithCursor(selectQuery, pagination, cursorParams, args...)

	batch.Queue(selectQuery, args...)

	connectionData = executePaginationBatch[*v2.RewardEdge, entities.Reward](ctx, &batch, rs.Connection, pagination)
	return connectionData
}

func (rs *Rewards) GetByOffset(ctx context.Context,
	partyIDHex *string,
	assetIDHex *string,
	pagination *entities.OffsetPagination,
) ([]entities.Reward, error) {
	query, _, args := selectRewards(partyIDHex, assetIDHex)

	if pagination != nil {
		order_cols := []string{"epoch_id", "party_id", "asset_id"}
		query, args = orderAndPaginateQuery(query, order_cols, *pagination, args...)
	}

	rewards := []entities.Reward{}
	defer metrics.StartSQLQuery("Rewards", "Get")()
	err := pgxscan.Select(ctx, rs.Connection, &rewards, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying rewards: %w", err)
	}
	return rewards, nil
}

func selectRewards(partyIDHex, assetIDHex *string) (string, string, []interface{}) {
	var args []interface{}
	var where string

	where, args = getRewardsWhereClause(partyIDHex, assetIDHex, args...)

	query := fmt.Sprintf(`SELECT * from rewards %s`, where)
	countQuery := fmt.Sprintf(`SELECT count(*) from rewards %s`, where)

	return query, countQuery, args
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
	where, newArgs := getRewardsWhereClause(partyIDHex, assetIDHex, *args...)

	*queryPtr = fmt.Sprintf("%s %s", query, where)
	*args = newArgs

	return nil
}

func getRewardsWhereClause(partyIDHex, assetIDHex *string, args ...interface{}) (string, []interface{}) {
	var where string

	if partyIDHex != nil && *partyIDHex != "" {
		partyID := entities.NewPartyID(*partyIDHex)
		where = fmt.Sprintf("WHERE party_id=%s", nextBindVar(&args, partyID))
	}

	if assetIDHex != nil && *assetIDHex != "" {
		clause := "WHERE"
		if partyIDHex != nil {
			clause = "AND"
		}

		assetID := entities.ID(*assetIDHex)
		where = fmt.Sprintf("%s %s asset_id=%s", where, clause, nextBindVar(&args, assetID))
	}

	return where, args
}
