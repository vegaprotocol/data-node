package sqlstore

import (
	"context"
	"fmt"
	"strings"

	"code.vegaprotocol.io/data-node/entities"
	"github.com/georgysavva/scany/pgxscan"
)

type Proposals struct {
	*SQLStore
}

func NewProposals(sqlStore *SQLStore) *Proposals {
	p := &Proposals{
		SQLStore: sqlStore,
	}
	return p
}

func (ps *Proposals) Add(ctx context.Context, r entities.Proposal) error {
	_, err := ps.pool.Exec(ctx,
		`INSERT INTO proposals(
			id,
			reference,
			party_id,
			state,
			terms,
			reason,
			error_details,
			proposal_time,
			vega_time)
		 VALUES ($1,  $2,  $3,  $4,  $5,  $6, $7, $8, $9)
		 ON CONFLICT (id, vega_time) DO UPDATE SET
			reference = EXCLUDED.reference,
			party_id = EXCLUDED.party_id,
			state = EXCLUDED.state,
			terms = EXCLUDED.terms,
			reason = EXCLUDED.reason,
			error_details = EXCLUDED.error_details,
			proposal_time = EXCLUDED.proposal_time
			;
		 `,
		r.ID, r.Reference, r.PartyID, r.State, r.Terms, r.Reason, r.ErrorDetails, r.ProposalTime, r.VegaTime)
	return err
}

func (ps *Proposals) GetByID(ctx context.Context, ID string) (entities.Proposal, error) {
	idBytes, err := entities.MakeProposalID(ID)
	if err != nil {
		return entities.Proposal{}, err
	}

	var p entities.Proposal
	query := `SELECT * FROM proposals_current WHERE id=$1`
	err = pgxscan.Get(ctx, ps.pool, &p, query, idBytes)
	return p, err
}

func (ps *Proposals) GetByReference(ctx context.Context, ref string) (entities.Proposal, error) {
	var p entities.Proposal
	query := `SELECT * FROM proposals_current WHERE reference=$1 LIMIT 1`
	err := pgxscan.Get(ctx, ps.pool, &p, query, ref)
	return p, err
}

func (ps *Proposals) Get(ctx context.Context,
	inState *entities.ProposalState,
	partyIDHex *string,
	proposalType *entities.ProposalType,
) ([]entities.Proposal, error) {
	query := `SELECT * FROM proposals_current`
	args := []interface{}{}

	conditions := []string{}

	if inState != nil {
		conditions = append(conditions, fmt.Sprintf("state=%s", nextBindVar(&args, *inState)))
	}

	if partyIDHex != nil {
		partyID, err := entities.MakePartyID(*partyIDHex)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, fmt.Sprintf("party_id=%s", nextBindVar(&args, partyID)))
	}

	if proposalType != nil {
		conditions = append(conditions, fmt.Sprintf("terms ? %s", nextBindVar(&args, *proposalType)))
	}

	if len(conditions) > 0 {
		query = fmt.Sprintf("%s WHERE %s", query, strings.Join(conditions, " AND "))
	}

	proposals := []entities.Proposal{}
	err := pgxscan.Select(ctx, ps.pool, &proposals, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying proposals: %w", err)
	}
	return proposals, nil

}
