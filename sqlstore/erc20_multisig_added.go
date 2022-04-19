package sqlstore

import (
	"context"
	"fmt"

	"code.vegaprotocol.io/data-node/entities"
	"github.com/georgysavva/scany/pgxscan"
)

type ERC20MultiSigSignerAdded struct {
	*SQLStore
}

func NewERC20MultiSigSignerAdded(sqlStore *SQLStore) *ERC20MultiSigSignerAdded {
	return &ERC20MultiSigSignerAdded{
		SQLStore: sqlStore,
	}
}

func (m *ERC20MultiSigSignerAdded) Add(e *entities.ERC20MultiSigSignerAdded) error {
	ctx, cancel := context.WithTimeout(context.Background(), m.conf.Timeout.Duration)
	defer cancel()

	query := `INSERT INTO erc20_multisig_signer_added (id, validator_id, new_signer, submitter, nonce, timestamp, epoch_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING`

	if _, err := m.pool.Exec(ctx, query,
		e.ID,
		e.ValidatorID,
		e.NewSigner,
		e.Submitter,
		e.Nonce,
		e.Timestamp,
		e.EpochID,
	); err != nil {
		err = fmt.Errorf("could not insert multisig-signer-added into database: %w", err)
		return err
	}

	return nil
}

func (m *ERC20MultiSigSignerAdded) GetByValidatorID(ctx context.Context, validatorID string, epochID *int64, pagination entities.Pagination) ([]entities.ERC20MultiSigSignerAdded, error) {
	out := []entities.ERC20MultiSigSignerAdded{}
	prequery := `SELECT * FROM erc20_multisig_signer_added WHERE validator_id=$1`
	query, args := orderAndPaginateQuery(prequery, nil, pagination, entities.NewNodeID(validatorID))

	if epochID != nil {
		prequery += " AND epoch_id=$2"
		query, args = orderAndPaginateQuery(prequery, nil, pagination, entities.NewNodeID(validatorID), *epochID)
	}

	err := pgxscan.Select(ctx, m.pool, &out, query, args...)
	return out, err
}

func (m *ERC20MultiSigSignerAdded) GetByEpochID(ctx context.Context, epochID int64, pagination entities.Pagination) ([]entities.ERC20MultiSigSignerAdded, error) {
	out := []entities.ERC20MultiSigSignerAdded{}
	prequery := `SELECT * FROM erc20_multisig_signer_added WHERE epoch_id=$1`
	query, args := orderAndPaginateQuery(prequery, nil, pagination, epochID)
	err := pgxscan.Select(ctx, m.pool, &out, query, args...)
	return out, err
}
