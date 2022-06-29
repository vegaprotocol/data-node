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
	"sync"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/metrics"
	v2 "code.vegaprotocol.io/protos/data-node/api/v2"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type Assets struct {
	*ConnectionSource
	cache     map[string]entities.Asset
	cacheLock sync.Mutex
}

func NewAssets(connectionSource *ConnectionSource) *Assets {
	a := &Assets{
		ConnectionSource: connectionSource,
		cache:            make(map[string]entities.Asset),
	}
	return a
}

func (as *Assets) Add(ctx context.Context, a entities.Asset) error {
	defer metrics.StartSQLQuery("Assets", "Add")()
	_, err := as.Connection.Exec(ctx,
		`INSERT INTO assets(id, name, symbol, total_supply, decimals, quantum, source, erc20_contract, lifetime_limit, withdraw_threshold, vega_time)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		a.ID,
		a.Name,
		a.Symbol,
		a.TotalSupply,
		a.Decimals,
		a.Quantum,
		a.Source,
		a.ERC20Contract,
		a.LifetimeLimit,
		a.WithdrawThreshold,
		a.VegaTime)
	return err
}

func (as *Assets) GetByID(ctx context.Context, id string) (entities.Asset, error) {
	as.cacheLock.Lock()
	defer as.cacheLock.Unlock()

	if asset, ok := as.cache[id]; ok {
		return asset, nil
	}

	a := entities.Asset{}

	defer metrics.StartSQLQuery("Assets", "GetByID")()
	err := pgxscan.Get(ctx, as.Connection, &a,
		`SELECT id, name, symbol, total_supply, decimals, quantum, source, erc20_contract, lifetime_limit, withdraw_threshold, vega_time
		 FROM assets WHERE id=$1`,
		entities.NewAssetID(id))

	if err == nil {
		as.cache[id] = a
	}
	return a, err
}

func (as *Assets) GetAll(ctx context.Context) ([]entities.Asset, error) {
	assets := []entities.Asset{}
	defer metrics.StartSQLQuery("Assets", "GetAll")()
	query, _ := getAssetQuery()
	err := pgxscan.Select(ctx, as.Connection, &assets, query)
	return assets, err
}

func (as *Assets) GetAllWithCursorPagination(ctx context.Context, pagination entities.CursorPagination) entities.ConnectionData[*v2.AssetEdge, entities.Asset] {
	var args []interface{}

	sorting, cmp, cursor := extractPaginationInfo(pagination)

	cursorParams := []CursorQueryParameter{
		NewCursorQueryParameter("id", sorting, cmp, entities.NewAssetID(cursor)),
	}

	batch := pgx.Batch{}
	query, countQuery := getAssetQuery()

	batch.Queue(countQuery, args...)

	query, args = orderAndPaginateWithCursor(query, pagination, cursorParams, args...)

	batch.Queue(query, args...)

	defer metrics.StartSQLQuery("Assets", "GetAllWithCursorPagination")()

	connectionData := executePaginationBatch[*v2.AssetEdge, entities.Asset](ctx, &batch, as.Connection, pagination)
	return connectionData
}

func getAssetQuery() (string, string) {
	selectQuery := `SELECT id, name, symbol, total_supply, decimals, quantum, source, erc20_contract, lifetime_limit, withdraw_threshold, vega_time
		FROM assets`
	countQuery := `SELECT COUNT(*) FROM assets`

	return selectQuery, countQuery
}
