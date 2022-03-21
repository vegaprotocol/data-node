package sqlstore

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"code.vegaprotocol.io/data-node/entities"
	"github.com/georgysavva/scany/pgxscan"
)

type Withdrawals struct {
	*SQLStore
}

func NewWithdrawals(sqlStore *SQLStore) *Withdrawals {
	return &Withdrawals{
		SQLStore: sqlStore,
	}
}

func (w *Withdrawals) Upsert(withdrawal *entities.Withdrawal) error {
	ctx, cancel := context.WithTimeout(context.Background(), w.conf.Timeout.Duration)
	defer cancel()

	query := `select 1`

	if _, err := w.pool.Exec(ctx, query); err != nil {
		err = fmt.Errorf("could not insert withdrawal into database: %w", err)
		return err
	}

	return nil
}

func (w *Withdrawals) GetByID(ctx context.Context, withdrawalID string) (entities.Withdrawal, error) {
	id, err := hex.DecodeString(withdrawalID)
	if err != nil {
		return entities.Withdrawal{}, err
	}

	var withdrawal entities.Withdrawal

	query := `select *
		from withdrawals
		where id = $1
		order by id, party_id, vega_time desc`

	err = pgxscan.Get(ctx, w.pool, &withdrawal, query, id)
	return withdrawal, err
}

func (w *Withdrawals) GetByParty(ctx context.Context, partyID string, pagination entities.Pagination) []entities.Withdrawal {
	id, err := hex.DecodeString(partyID)
	if err != nil {
		return nil
	}

	var withdrawals []entities.Withdrawal
	var args []interface{}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(fmt.Sprintf(`select *
		from withdrawals
		where party_id = %s `, nextBindVar(&args, id)))

	queryBuilder.WriteString(" order by id, party_id, vega_time desc")

	var query string
	query, args = orderAndPaginateQuery(queryBuilder.String(), nil, pagination, args...)

	if err = pgxscan.Select(ctx, w.pool, &withdrawals, query, args...); err != nil {
		return nil
	}

	return withdrawals
}
