package sqlstore

import (
	"context"
	"errors"
	"fmt"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/vega/types/num"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/shopspring/decimal"
)

type StakeLinking struct {
	*SQLStore
}

const (
	sqlStakeLinkingColumns = `id, stake_linking_type, ethereum_timestamp, party_id, amount, stake_linking_status, finalized_at,
tx_hash, log_index, ethereum_address, vega_time`
)

func NewStakeLinking(sqlStore *SQLStore) *StakeLinking {
	return &StakeLinking{
		SQLStore: sqlStore,
	}
}

func (s *StakeLinking) Upsert(stake *entities.StakeLinking) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.conf.Timeout.Duration)
	defer cancel()

	query := fmt.Sprintf(`insert into stake_linking (%s)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
on conflict (id, vega_time) do update
set
	stake_linking_type=EXCLUDED.stake_linking_type,
	ethereum_timestamp=EXCLUDED.ethereum_timestamp,
	party_id=EXCLUDED.party_id,
	amount=EXCLUDED.amount,
	stake_linking_status=EXCLUDED.stake_linking_status,
	finalized_at=EXCLUDED.finalized_at,
	tx_hash=EXCLUDED.tx_hash,
	log_index=EXCLUDED.log_index,
	ethereum_address=EXCLUDED.ethereum_address`, sqlStakeLinkingColumns)

	if _, err := s.pool.Exec(ctx, query, stake.ID, stake.StakeLinkingType, stake.EthereumTimestamp, stake.PartyID, stake.Amount,
		stake.StakeLinkingStatus, stake.FinalizedAt, stake.TxHash, stake.LogIndex,
		stake.EthereumAddress, stake.VegaTime); err != nil {
		return err
	}

	return nil
}

func (s *StakeLinking) GetStake(ctx context.Context, partyID entities.PartyID,
	pagination entities.Pagination) (*num.Uint, []entities.StakeLinking) {
	var links []entities.StakeLinking
	var bindVars []interface{}
	// get the links from the database
	query := fmt.Sprintf(`select %s
from stake_linking_current
where party_id=%s`, sqlStakeLinkingColumns, nextBindVar(&bindVars, partyID))

	query, bindVars = orderAndPaginateQuery(query, nil, pagination, bindVars...)
	var bal *num.Uint
	var err error

	err = pgxscan.Select(ctx, s.pool, &links, query, bindVars...)
	if err != nil {
		s.log.Errorf("could not retrieve links", logging.Error(err))
		return bal, nil
	}

	bal, err = s.calculateBalance(ctx, partyID)
	if err != nil {
		s.log.Errorf("cannot calculate balance", logging.Error(err))
		return num.Zero(), nil
	}
	return bal, links
}

func (s *StakeLinking) calculateBalance(ctx context.Context, partyID entities.PartyID) (*num.Uint, error) {
	bal := num.Zero()
	var bindVars []interface{}

	query := fmt.Sprintf(`with cte_stake_linking(%s) as (
	select %s
	from stake_linking_current
	where party_id = %s
), ctelinks(party_id, amount) as (
    select party_id, sum(amount)
    from cte_stake_linking
    where stake_linking_type = 'TYPE_LINK'
    and stake_linking_status = 'STATUS_ACCEPTED'
    group by party_id
), cteunlinks(party_id, amount) as (
    select party_id, sum(amount)
    from cte_stake_linking
    where stake_linking_type = 'TYPE_UNLINK'
    and stake_linking_status = 'STATUS_ACCEPTED'
    group by party_id
), cteparty(party_id) as (
	-- this is to ensure we always return one row with the party_id that has been requested, even if we have no data
	select %s::bytea 
)
    select p.party_id, coalesce(l.amount, 0) - coalesce(u.amount, 0) as current_balance
    from cteparty p 
		full outer join ctelinks l on p.party_id = l.party_id
        full outer join cteunlinks u on l.party_id = u.party_id
`, sqlStakeLinkingColumns, sqlStakeLinkingColumns, nextBindVar(&bindVars, partyID), nextBindVar(&bindVars, partyID))

	result := struct {
		PartyID        []byte
		CurrentBalance decimal.Decimal
	}{}

	if err := pgxscan.Get(ctx, s.pool, &result, query, bindVars...); err != nil {
		return bal, err
	}

	if result.CurrentBalance.LessThan(decimal.Zero) {
		return bal, errors.New("unlinked amount is greater than linked amount, potential missed events")
	}

	var overflowed bool

	if bal, overflowed = num.UintFromDecimal(result.CurrentBalance); overflowed {
		return num.Zero(), fmt.Errorf("current balance is invalid: %s", result.CurrentBalance.String())
	}

	return bal, nil
}
