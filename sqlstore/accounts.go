package sqlstore

import (
	"context"
	"fmt"

	"code.vegaprotocol.io/data-node/entities"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
)

type Accounts struct {
	*SqlStore
}

func NewAccounts(sqlStore *SqlStore) *Accounts {
	a := &Accounts{
		SqlStore: sqlStore,
	}
	return a
}

// Updates the account supplied with autogenerated ID
func (as *Accounts) Add(a *entities.Account) error {
	ctx := context.Background()
	err := as.pool.QueryRow(ctx,
		`INSERT INTO accounts(party_id, asset_id, market_id, type, vega_time)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		a.PartyID,
		a.AssetID,
		a.MarketID,
		a.Type,
		a.VegaTime).Scan(&a.ID)
	return err
}

func (as *Accounts) GetByID(id int64) (entities.Account, error) {
	a := entities.Account{}
	ctx := context.Background()
	err := pgxscan.Get(ctx, as.pool, &a,
		`SELECT id, party_id, asset_id, market_id, type, vega_time
		 FROM accounts WHERE id=$1`,
		id)
	return a, err
}

func (as *Accounts) GetAll() ([]entities.Account, error) {
	ctx := context.Background()
	accounts := []entities.Account{}
	err := pgxscan.Select(ctx, as.pool, &accounts, `
		SELECT id, party_id, asset_id, market_id, type, vega_time
		FROM accounts`)
	return accounts, err
}

// If an account with matching party/asset/market/type does not exist in the database, create one.
// If an account already exists, fetch that one.
// In either case, update the entities.Account object passed with an ID from the database.
func (as *Accounts) Obtain(a *entities.Account) error {
	insertQuery := `INSERT INTO accounts(party_id, asset_id, market_id, type, vega_time)
                           VALUES ($1, $2, $3, $4, $5)
                           ON CONFLICT (party_id, asset_id, market_id, type) DO NOTHING`

	selectQuery := `SELECT id, party_id, asset_id, market_id, type, vega_time
	                FROM accounts
	                WHERE party_id=$1 AND asset_id=$2 AND market_id=$3 AND type=$4`

	batch := pgx.Batch{}
	batch.Queue(insertQuery, a.PartyID, a.AssetID, a.MarketID, a.Type, a.VegaTime)
	batch.Queue(selectQuery, a.PartyID, a.AssetID, a.MarketID, a.Type)
	results := as.pool.SendBatch(context.Background(), &batch)
	defer results.Close()

	if _, err := results.Exec(); err != nil {
		return fmt.Errorf("inserting account: %w", err)
	}

	rows, err := results.Query()
	if err != nil {
		return fmt.Errorf("querying accounts: %w", err)
	}

	if err = pgxscan.ScanOne(a, rows); err != nil {
		return fmt.Errorf("scanning account: %w", err)
	}

	return nil
}
