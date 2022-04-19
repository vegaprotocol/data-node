package sqlstore

import (
	"context"
	"fmt"

	"code.vegaprotocol.io/data-node/entities"
	"github.com/georgysavva/scany/pgxscan"
)

type ERC20MultiSigSignerRemoved struct {
	*SQLStore
}

func NewERC20MultiSigSignerRemoved(sqlStore *SQLStore) *ERC20MultiSigSignerRemoved {
	return &ERC20MultiSigSignerRemoved{
		SQLStore: sqlStore,
	}
}

func (m *ERC20MultiSigSignerRemoved) Add(e *entities.ERC20MultiSigSignerRemoved) error {
	ctx, cancel := context.WithTimeout(context.Background(), m.conf.Timeout.Duration)
	defer cancel()

	query := `INSERT INTO erc20_multisig_signer_removed (id, validator_id, old_signer, submitter, nonce, timestamp, epoch_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING`

	if _, err := m.pool.Exec(ctx, query,
		e.ID,
		e.ValidatorID,
		e.OldSigner,
		e.Submitter,
		e.Nonce,
		e.Timestamp,
		e.EpochID,
	); err != nil {
		err = fmt.Errorf("could not insert multisig-signer-removed into database: %w", err)
		return err
	}

	return nil
}

func (m *ERC20MultiSigSignerRemoved) GetByValidatorID(ctx context.Context, validatorID string, submitter string, epochID *int64, pagination entities.Pagination) ([]entities.ERC20MultiSigSignerRemoved, error) {
	out := []entities.ERC20MultiSigSignerRemoved{}
	prequery := `SELECT * FROM erc20_multisig_signer_removed WHERE validator_id=$1 AND submitter=$2`
	query, args := orderAndPaginateQuery(prequery, nil, pagination, entities.NewNodeID(validatorID), entities.NewEthereumAddress(submitter))

	if epochID != nil {
		prequery += " AND epoch_id=$3"
		query, args = orderAndPaginateQuery(prequery, nil, pagination, entities.NewNodeID(validatorID), entities.NewEthereumAddress(submitter), *epochID)
	}
	err := pgxscan.Select(ctx, m.pool, &out, query, args...)
	return out, err
}
