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

type OracleData struct {
	*ConnectionSource
}

const (
	sqlOracleDataColumns = `public_keys, data, matched_spec_ids, broadcast_at, vega_time`
)

func NewOracleData(connectionSource *ConnectionSource) *OracleData {
	return &OracleData{
		ConnectionSource: connectionSource,
	}
}

func (od *OracleData) Add(ctx context.Context, data *entities.OracleData) error {
	defer metrics.StartSQLQuery("OracleData", "Add")()
	query := fmt.Sprintf("insert into oracle_data(%s) values ($1, $2, $3, $4, $5)", sqlOracleDataColumns)

	if _, err := od.Connection.Exec(ctx, query, data.PublicKeys, data.Data, data.MatchedSpecIds, data.BroadcastAt, data.VegaTime); err != nil {
		err = fmt.Errorf("could not insert oracle data into database: %w", err)
		return err
	}

	return nil
}

func (od *OracleData) GetOracleDataBySpecID(ctx context.Context, id string, pagination entities.Pagination) entities.ConnectionData[*v2.OracleDataEdge, entities.OracleData] {
	switch p := pagination.(type) {
	case entities.OffsetPagination:
		return getOracleDataBySpecIDOffsetPagination(ctx, od.Connection, id, p)
	case entities.CursorPagination:
		return getOracleDataBySpecIDCursorPagination(ctx, od.Connection, id, p)
	default:
		return entities.ConnectionData[*v2.OracleDataEdge, entities.OracleData]{
			Err: fmt.Errorf("unrecognised pagination: %v", p),
		}
	}
}

func getOracleDataBySpecIDOffsetPagination(ctx context.Context, conn Connection, id string, pagination entities.OffsetPagination) entities.ConnectionData[*v2.OracleDataEdge, entities.OracleData] {
	specID := entities.NewSpecID(id)
	var bindVars []interface{}

	query := fmt.Sprintf(`select %s
	from oracle_data where %s = ANY(matched_spec_ids)`, sqlOracleDataColumns, nextBindVar(&bindVars, specID))

	query, bindVars = orderAndPaginateQuery(query, nil, pagination, bindVars...)
	var oracleData []entities.OracleData

	defer metrics.StartSQLQuery("OracleData", "GetBySpecID")()
	err := pgxscan.Select(ctx, conn, &oracleData, query, bindVars...)

	if err != nil {
		connectionData := entities.ConnectionData[*v2.OracleDataEdge, entities.OracleData]{
			Err: err,
		}

		return connectionData
	}

	connectionData := entities.ConnectionData[*v2.OracleDataEdge, entities.OracleData]{
		TotalCount: int64(len(oracleData)),
		Entities:   oracleData,
		PageInfo: entities.PageInfo{
			HasNextPage:     false,
			HasPreviousPage: false,
			StartCursor:     oracleData[0].Cursor().Encode(),
			EndCursor:       oracleData[len(oracleData)-1].Cursor().Encode(),
		},
	}

	return connectionData
}

func getOracleDataBySpecIDCursorPagination(ctx context.Context, conn Connection, id string, pagination entities.CursorPagination) entities.ConnectionData[*v2.OracleDataEdge, entities.OracleData] {
	var bindVars []interface{}

	specID := entities.NewSpecID(id)
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(fmt.Sprintf(`select %s from oracle_data`, sqlOracleDataColumns))

	countBuilder := strings.Builder{}
	countBuilder.WriteString(`select count(*) from oracle_data`)

	where := fmt.Sprintf(` where %s = ANY(matched_spec_ids)`, nextBindVar(&bindVars, specID))

	queryBuilder.WriteString(where)
	countBuilder.WriteString(where)

	batch := pgx.Batch{}
	batch.Queue(countBuilder.String(), bindVars...)

	sorting, cmp, cursor := extractPaginationInfo(pagination)

	dc := &entities.OracleDataCursor{}
	if cursor != "" {
		err := dc.Parse(cursor)
		if err != nil {
			return entities.ConnectionData[*v2.OracleDataEdge, entities.OracleData]{
				Err: fmt.Errorf("parsing cursor information: %w", err),
			}
		}
	}

	cursorParams := []CursorQueryParameter{
		NewCursorQueryParameter("vega_time", sorting, cmp, dc.VegaTime),
		NewCursorQueryParameter("matched_spec_ids", sorting, cmp, nil),
	}

	query := queryBuilder.String()
	query, bindVars = orderAndPaginateWithCursor(query, pagination, cursorParams, bindVars...)

	batch.Queue(query, bindVars...)

	defer metrics.StartSQLQuery("OracleData", "ListOracleData")()

	connectionData := executePaginationBatch[*v2.OracleDataEdge, entities.OracleData](ctx, &batch, conn, pagination)
	return connectionData
}

func (od *OracleData) ListOracleData(ctx context.Context, pagination entities.Pagination) entities.ConnectionData[*v2.OracleDataEdge, entities.OracleData] {
	switch p := pagination.(type) {
	case entities.OffsetPagination:
		return listOracleDataOffsetPagination(ctx, od.Connection, p)
	case entities.CursorPagination:
		return listOracleDataCursorPagination(ctx, od.Connection, p)
	default:
		return entities.ConnectionData[*v2.OracleDataEdge, entities.OracleData]{Err: fmt.Errorf("unrecognised pagination: %v", p)}
	}
}

func listOracleDataOffsetPagination(ctx context.Context, conn Connection, pagination entities.OffsetPagination) entities.ConnectionData[*v2.OracleDataEdge, entities.OracleData] {

	var data []entities.OracleData
	query, _ := selectOracleData()
	query = fmt.Sprintf(`%s
order by vega_time desc, matched_spec_id`, query)

	var bindVars []interface{}
	query, bindVars = orderAndPaginateQuery(query, nil, pagination, bindVars...)
	defer metrics.StartSQLQuery("OracleData", "ListOracleData")()
	err := pgxscan.Select(ctx, conn, &data, query, bindVars...)

	if err != nil {
		return entities.ConnectionData[*v2.OracleDataEdge, entities.OracleData]{
			Err: err,
		}
	}

	connectionData := entities.ConnectionData[*v2.OracleDataEdge, entities.OracleData]{
		TotalCount: int64(len(data)),
		Entities:   data,
		PageInfo: entities.PageInfo{
			HasNextPage:     false,
			HasPreviousPage: false,
			StartCursor:     data[0].Cursor().Encode(),
			EndCursor:       data[len(data)-1].Cursor().Encode(),
		},
		Err: nil,
	}

	return connectionData
}

func listOracleDataCursorPagination(ctx context.Context, conn Connection, pagination entities.CursorPagination) entities.ConnectionData[*v2.OracleDataEdge, entities.OracleData] {
	query, countQuery := selectOracleData()
	var bindVars []interface{}

	batch := pgx.Batch{}
	batch.Queue(countQuery, bindVars...)

	sorting, cmp, cursor := extractPaginationInfo(pagination)

	dc := &entities.OracleDataCursor{}
	if cursor != "" {
		if err := dc.Parse(cursor); err != nil {
			return entities.ConnectionData[*v2.OracleDataEdge, entities.OracleData]{
				Err: fmt.Errorf("parsing cursor information: %w", err),
			}
		}
	}

	cursorParams := []CursorQueryParameter{
		NewCursorQueryParameter("vega_time", sorting, cmp, dc.VegaTime),
		NewCursorQueryParameter("matched_spec_ids", sorting, cmp, nil),
	}

	query, bindVars = orderAndPaginateWithCursor(query, pagination, cursorParams, bindVars...)
	batch.Queue(query, bindVars...)

	connectionData := executePaginationBatch[*v2.OracleDataEdge, entities.OracleData](ctx, &batch, conn, pagination)
	return connectionData
}

func selectOracleData() (string, string) {
	query := fmt.Sprintf(`select %s
from oracle_data_current`, sqlOracleDataColumns)
	countQuery := fmt.Sprintf(`select count(*)
from oracle_data_current`)
	return query, countQuery
}
