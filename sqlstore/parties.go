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

	v2 "code.vegaprotocol.io/protos/data-node/api/v2"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/metrics"
	"github.com/georgysavva/scany/pgxscan"
)

var (
	ErrPartyNotFound  = errors.New("party not found")
	ErrInvalidPartyID = errors.New("invalid hex id")
)

type Parties struct {
	*ConnectionSource
}

func NewParties(connectionSource *ConnectionSource) *Parties {
	ps := &Parties{
		ConnectionSource: connectionSource,
	}
	return ps
}

// Initialise adds the built-in 'network' party which is never explicitly sent on the event
// bus, but nonetheless is necessary.
func (ps *Parties) Initialise() {
	defer metrics.StartSQLQuery("Parties", "Initialise")()
	_, err := ps.Connection.Exec(context.Background(),
		`INSERT INTO parties(id) VALUES ($1) ON CONFLICT (id) DO NOTHING`,
		entities.NewPartyID("network"))
	if err != nil {
		panic(fmt.Errorf("Unable to add built-in network party: %w", err))
	}
}

func (ps *Parties) Add(ctx context.Context, p entities.Party) error {
	defer metrics.StartSQLQuery("Parties", "Add")()
	_, err := ps.Connection.Exec(ctx,
		`INSERT INTO parties(id, vega_time)
		 VALUES ($1, $2)
		 ON CONFLICT (id) DO NOTHING`,
		p.ID,
		p.VegaTime)
	return err
}

func (ps *Parties) GetByID(ctx context.Context, id string) (entities.Party, error) {
	a := entities.Party{}
	defer metrics.StartSQLQuery("Parties", "GetByID")()
	err := pgxscan.Get(ctx, ps.Connection, &a,
		`SELECT id, vega_time
		 FROM parties WHERE id=$1`,
		entities.NewPartyID(id))

	if pgxscan.NotFound(err) {
		return a, fmt.Errorf("'%v': %w", id, ErrPartyNotFound)
	}

	if errors.Is(err, entities.ErrInvalidID) {
		return a, fmt.Errorf("'%v': %w", id, ErrInvalidPartyID)
	}

	return a, err
}

func (ps *Parties) GetAll(ctx context.Context) ([]entities.Party, error) {
	parties := []entities.Party{}
	defer metrics.StartSQLQuery("Parties", "GetAll")()
	err := pgxscan.Select(ctx, ps.Connection, &parties, `
		SELECT id, vega_time
		FROM parties`)
	return parties, err
}

func (ps *Parties) GetAllPaged(ctx context.Context, partyID string, pagination entities.CursorPagination) entities.ConnectionData[*v2.PartyEdge, entities.Party] {
	var connectionData entities.ConnectionData[*v2.PartyEdge, entities.Party]

	if partyID != "" {
		party, err := ps.GetByID(ctx, partyID)
		if err != nil {
			connectionData.Err = err
			return connectionData
		}

		connectionData.TotalCount = 1
		connectionData.Entities = append(connectionData.Entities, party)
		connectionData.PageInfo = entities.PageInfo{
			StartCursor: party.Cursor().Encode(),
			EndCursor:   party.Cursor().Encode(),
		}

		return connectionData
	}

	args := make([]interface{}, 0)

	countQuery := `SELECT COUNT(*) FROM parties`

	batch := pgx.Batch{}
	batch.Queue(countQuery)

	query := `
		SELECT id, vega_time
		FROM parties
	`

	sorting, cmp, cursor := extractPaginationInfo(pagination)
	cursors := []CursorQueryParameter{NewCursorQueryParameter("vega_time", sorting, cmp, cursor)}
	query, args = orderAndPaginateWithCursor(query, pagination, cursors, args...)

	batch.Queue(query, args...)

	connectionData = executePaginationBatch[*v2.PartyEdge, entities.Party](ctx, &batch, ps.Connection, pagination)
	return connectionData
}
