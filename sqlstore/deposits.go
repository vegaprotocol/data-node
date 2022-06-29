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
	"strings"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/metrics"
	v2 "code.vegaprotocol.io/protos/data-node/api/v2"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type Deposits struct {
	*ConnectionSource
}

const (
	sqlDepositsColumns = `id, status, party_id, asset, amount, tx_hash,
		credited_timestamp, created_timestamp, vega_time`
)

func NewDeposits(connectionSource *ConnectionSource) *Deposits {
	return &Deposits{
		ConnectionSource: connectionSource,
	}
}

func (d *Deposits) Upsert(ctx context.Context, deposit *entities.Deposit) error {
	query := fmt.Sprintf(`insert into deposits(%s)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
on conflict (id, party_id, vega_time) do update
set
	status=EXCLUDED.status,
	asset=EXCLUDED.asset,
	amount=EXCLUDED.amount,
	tx_hash=EXCLUDED.tx_hash,
	credited_timestamp=EXCLUDED.credited_timestamp,
	created_timestamp=EXCLUDED.created_timestamp`, sqlDepositsColumns)

	defer metrics.StartSQLQuery("Deposits", "Upsert")()
	if _, err := d.Connection.Exec(ctx, query, deposit.ID, deposit.Status, deposit.PartyID, deposit.Asset, deposit.Amount,
		deposit.TxHash, deposit.CreditedTimestamp, deposit.CreatedTimestamp, deposit.VegaTime); err != nil {
		err = fmt.Errorf("could not insert deposit into database: %w", err)
		return err
	}

	return nil
}

func (d *Deposits) GetByID(ctx context.Context, depositID string) (entities.Deposit, error) {
	var deposit entities.Deposit

	query := `select id, status, party_id, asset, amount, tx_hash, credited_timestamp, created_timestamp, vega_time
		from deposits_current
		where id = $1
		order by id, party_id, vega_time desc`

	defer metrics.StartSQLQuery("Deposits", "GetByID")()
	err := pgxscan.Get(ctx, d.Connection, &deposit, query, entities.NewDepositID(depositID))
	return deposit, err
}

func (d *Deposits) GetByParty(ctx context.Context, party string, openOnly bool, pagination entities.Pagination) entities.ConnectionData[*v2.DepositEdge, entities.Deposit] {
	switch p := pagination.(type) {
	case entities.OffsetPagination:
		return d.getByPartyOffsetPagination(ctx, party, openOnly, p)
	case entities.CursorPagination:
		return d.getByPartyCursorPagination(ctx, party, openOnly, p)
	default:
		return d.getByPartyOffsetPagination(ctx, party, openOnly, entities.OffsetPagination{})
	}
}

func (d *Deposits) getByPartyOffsetPagination(ctx context.Context, party string, openOnly bool,
	pagination entities.OffsetPagination) entities.ConnectionData[*v2.DepositEdge, entities.Deposit] {
	var connectionData entities.ConnectionData[*v2.DepositEdge, entities.Deposit]

	query, _, args := getDepositsByPartyQuery(party)

	if openOnly {
		query = fmt.Sprintf(`%s and status = %s`, query, nextBindVar(&args, entities.DepositStatusOpen))
	}
	query = fmt.Sprintf("%s order by id, party_id, vega_time desc", query)
	query, args = orderAndPaginateQuery(query, nil, pagination, args...)

	defer metrics.StartSQLQuery("Deposits", "GetByParty")()
	if err := pgxscan.Select(ctx, d.Connection, &connectionData.Entities, query, args...); err != nil {
		connectionData.Err = fmt.Errorf("could not get deposits by party: %w", err)
		return connectionData
	}

	connectionData.TotalCount = int64(len(connectionData.Entities))
	connectionData.PageInfo = entities.PageInfo{
		HasNextPage:     false,
		HasPreviousPage: false,
		StartCursor:     connectionData.Entities[0].Cursor().Encode(),
		EndCursor:       connectionData.Entities[len(connectionData.Entities)-1].Cursor().Encode(),
	}

	return connectionData
}

func (d *Deposits) getByPartyCursorPagination(ctx context.Context, party string, openOnly bool,
	pagination entities.CursorPagination) entities.ConnectionData[*v2.DepositEdge, entities.Deposit] {
	var connectionData entities.ConnectionData[*v2.DepositEdge, entities.Deposit]

	sorting, cmp, cursor := extractPaginationInfo(pagination)

	dc := &entities.DepositCursor{}
	if err := dc.Parse(cursor); err != nil {
		connectionData.Err = fmt.Errorf("could not parse cursor information: %w", err)
		return connectionData
	}

	cursorParams := []CursorQueryParameter{
		NewCursorQueryParameter("vega_time", sorting, cmp, dc.VegaTime),
		NewCursorQueryParameter("id", sorting, cmp, entities.NewDepositID(dc.ID)),
	}

	selectQuery, countQuery, args := getDepositsByPartyQuery(party)

	batch := pgx.Batch{}

	if openOnly {
		and := fmt.Sprintf("and status = %s", nextBindVar(&args, entities.DepositStatusOpen))

		selectQuery = fmt.Sprintf(`%s %s`, selectQuery, and)
		countQuery = fmt.Sprintf(`%s %s`, countQuery, and)
	}

	batch.Queue(countQuery, args...)

	selectQuery, args = orderAndPaginateWithCursor(selectQuery, pagination, cursorParams, args...)
	batch.Queue(selectQuery, args...)

	connectionData = executePaginationBatch[*v2.DepositEdge, entities.Deposit](ctx, &batch, d.Connection, pagination)
	return connectionData
}

func getDepositsByPartyQuery(partyID string) (string, string, []interface{}) {
	var args []interface{}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`select id, status, party_id, asset, amount, tx_hash, credited_timestamp, created_timestamp, vega_time
		from deposits_current`)

	countBuilder := strings.Builder{}
	countBuilder.WriteString(`select count(*) from deposits_current`)

	where := fmt.Sprintf(" where party_id = %s", nextBindVar(&args, entities.NewPartyID(partyID)))

	queryBuilder.WriteString(where)
	countBuilder.WriteString(where)

	return queryBuilder.String(), countBuilder.String(), args
}
