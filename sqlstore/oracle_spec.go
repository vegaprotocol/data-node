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

type OracleSpec struct {
	*ConnectionSource
}

const (
	sqlOracleSpecColumns = `id, created_at, updated_at, public_keys, filters, status, vega_time`
)

func NewOracleSpec(connectionSource *ConnectionSource) *OracleSpec {
	return &OracleSpec{
		ConnectionSource: connectionSource,
	}
}

func (os *OracleSpec) Upsert(ctx context.Context, spec *entities.OracleSpec) error {
	query := fmt.Sprintf(`insert into oracle_specs(%s)
values ($1, $2, $3, $4, $5, $6, $7)
on conflict (id, vega_time) do update
set
	created_at=EXCLUDED.created_at,
	updated_at=EXCLUDED.updated_at,
	public_keys=EXCLUDED.public_keys,
	filters=EXCLUDED.filters,
	status=EXCLUDED.status`, sqlOracleSpecColumns)

	defer metrics.StartSQLQuery("OracleSpec", "Upsert")()
	if _, err := os.Connection.Exec(ctx, query, spec.ID, spec.CreatedAt, spec.UpdatedAt, spec.PublicKeys,
		spec.Filters, spec.Status, spec.VegaTime); err != nil {
		return err
	}

	return nil
}

func (os *OracleSpec) GetSpecByID(ctx context.Context, specID string) (entities.OracleSpec, error) {
	var spec entities.OracleSpec
	query := fmt.Sprintf(`%s
where id = $1
order by id, vega_time desc`, getOracleSpecsQuery())

	defer metrics.StartSQLQuery("OracleSpec", "GetByID")()
	err := pgxscan.Get(ctx, os.Connection, &spec, query, entities.NewSpecID(specID))
	return spec, err
}

func (os *OracleSpec) GetSpecs(ctx context.Context, pagination entities.OffsetPagination) ([]entities.OracleSpec, error) {
	var specs []entities.OracleSpec
	query := fmt.Sprintf(`%s order by id, vega_time desc`, getOracleSpecsQuery())

	var bindVars []interface{}
	query, bindVars = orderAndPaginateQuery(query, nil, pagination, bindVars...)
	defer metrics.StartSQLQuery("OracleSpec", "ListOracleSpecs")()
	err := pgxscan.Select(ctx, os.Connection, &specs, query, bindVars...)
	return specs, err
}

func (os *OracleSpec) GetSpecsWithCursorPagination(ctx context.Context, specID string, pagination entities.CursorPagination) entities.ConnectionData[*v2.OracleSpecEdge, entities.OracleSpec] {
	if specID != "" {
		return os.getSingleSpecWithPageInfo(ctx, specID)
	}

	return os.getSpecsWithPageInfo(ctx, pagination)
}

func (os *OracleSpec) getSingleSpecWithPageInfo(ctx context.Context, specID string) entities.ConnectionData[*v2.OracleSpecEdge, entities.OracleSpec] {
	var connectionData entities.ConnectionData[*v2.OracleSpecEdge, entities.OracleSpec]
	spec, err := os.GetSpecByID(ctx, specID)
	if err != nil {
		connectionData.Err = err
		return connectionData
	}

	connectionData.TotalCount = 1
	connectionData.Entities = []entities.OracleSpec{
		spec,
	}
	connectionData.PageInfo = entities.PageInfo{
		HasNextPage:     false,
		HasPreviousPage: false,
		StartCursor:     spec.Cursor().Encode(),
		EndCursor:       spec.Cursor().Encode(),
	}

	return connectionData
}

func (os *OracleSpec) getSpecsWithPageInfo(ctx context.Context, pagination entities.CursorPagination) entities.ConnectionData[*v2.OracleSpecEdge, entities.OracleSpec] {
	sorting, cmp, cursor := extractPaginationInfo(pagination)
	cursorParams := []CursorQueryParameter{
		NewCursorQueryParameter("id", sorting, cmp, entities.NewSpecID(cursor)),
		NewCursorQueryParameter("vega_time", "desc", cmp, nil),
	}

	var args []interface{}
	query := getOracleSpecsQuery()
	countQuery := fmt.Sprintf(`select count(sp.*) from (%s) sp`, query)

	batch := pgx.Batch{}
	batch.Queue(countQuery, args...)

	query, args = orderAndPaginateWithCursor(query, pagination, cursorParams, args...)
	batch.Queue(query, args...)

	connectionData := executePaginationBatch[*v2.OracleSpecEdge, entities.OracleSpec](ctx, &batch, os.Connection, pagination)
	return connectionData
}

func getOracleSpecsQuery() string {
	return fmt.Sprintf(`select distinct on (id) %s
from oracle_specs`, sqlOracleSpecColumns)
}
