package sqlstore

import (
	"context"

	"code.vegaprotocol.io/data-node/entities"
	"github.com/georgysavva/scany/pgxscan"
)

type Ledger struct {
	*SQLStore
}

func NewLedger(sqlStore *SQLStore) *Ledger {
	a := &Ledger{
		SQLStore: sqlStore,
	}
	return a
}

// Add inserts a record in the database and updates the supplied struct with the autogenerated ID
func (ls *Ledger) Add(le *entities.LedgerEntry) error {
	ctx := context.Background()
	err := ls.pool.QueryRow(ctx,
		`INSERT INTO ledger(account_from_id, account_to_id, quantity, vega_time, transfer_time, reference, type)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		le.AccountFromID,
		le.AccountToID,
		le.Quantity,
		le.VegaTime,
		le.TransferTime,
		le.Reference,
		le.Type).Scan(&le.ID)
	return err
}

func (ls *Ledger) GetByID(id int64) (entities.LedgerEntry, error) {
	le := entities.LedgerEntry{}
	ctx := context.Background()
	err := pgxscan.Get(ctx, ls.pool, &le,
		`SELECT id, account_from_id, account_to_id, quantity, vega_time, transfer_time, reference, type
		 FROM ledger WHERE id=$1`,
		id)
	return le, err
}

func (ls *Ledger) GetAll() ([]entities.LedgerEntry, error) {
	ctx := context.Background()
	ledgerEntries := []entities.LedgerEntry{}
	err := pgxscan.Select(ctx, ls.pool, &ledgerEntries, `
		SELECT id, account_from_id, account_to_id, quantity, vega_time, transfer_time, reference, type
		FROM ledger`)
	return ledgerEntries, err
}
