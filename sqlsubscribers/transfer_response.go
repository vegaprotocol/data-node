package sqlsubscribers

import (
	"context"
	"fmt"
	"time"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/data-node/subscribers"
	"code.vegaprotocol.io/protos/vega"
	"code.vegaprotocol.io/vega/events"
	"github.com/shopspring/decimal"
)

type Ledger interface {
	Add(*entities.LedgerEntry) error
}

type AccountStore interface {
	Obtain(a *entities.Account) error
}

type BalanceStore interface {
	Add(b entities.Balance) error
}
type PartyStore interface {
}

type TransferResponseEvent interface {
	events.Event
	TransferResponses() []*vega.TransferResponse
}

type TransferResponse struct {
	*subscribers.Base
	ledger   Ledger
	accounts AccountStore
	parties  PartyStore
	blocks   BlockStore
	balances BalanceStore
	log      *logging.Logger
}

func NewTransferResponse(
	ctx context.Context,
	ledger Ledger,
	accounts AccountStore,
	balances BalanceStore,
	parties PartyStore,
	blocks BlockStore,
	log *logging.Logger,
) *TransferResponse {
	return &TransferResponse{
		Base:     subscribers.NewBase(ctx, 0, true),
		ledger:   ledger,
		accounts: accounts,
		balances: balances,
		parties:  parties,
		blocks:   blocks,
		log:      log,
	}
}

func (t *TransferResponse) Types() []events.Type {
	return []events.Type{
		events.TransferResponses,
	}
}

func (t *TransferResponse) Push(evts ...events.Event) {
	for _, e := range evts {
		if tre, ok := e.(TransferResponseEvent); ok {
			t.consume(tre)
		}
	}
}

func (t *TransferResponse) consume(e TransferResponseEvent) {
	t.log.Debug("TransferResponseEvent: ", logging.Int64("block", e.BlockNr()))

	var err error
	block, err := t.blocks.WaitForBlockHeight(e.BlockNr())
	if err != nil {
		t.log.Error("can't ingest transfer response because we don't have block")
		return
	}

	for _, tr := range e.TransferResponses() {
		for _, vle := range tr.Transfers {
			if err := t.addLedgerEntry(vle, block.VegaTime); err != nil {
				t.log.Error("couldn't add ledger entry",
					logging.Error(err),
					logging.Reflect("ledgerEntry", vle))
			}
		}
		for _, vb := range tr.Balances {
			if err := t.addBalance(vb, block.VegaTime); err != nil {
				t.log.Error("couldn't add balance",
					logging.Error(err),
					logging.Reflect("balance", vb))
			}
		}
	}
}

func (t *TransferResponse) addBalance(vb *vega.TransferBalance, vegaTime time.Time) error {
	acc, err := t.obtainAccountWithProto(vb.Account, vegaTime)
	if err != nil {
		return fmt.Errorf("obtaining account: %w", err)
	}

	balance, err := decimal.NewFromString(vb.Balance)
	if err != nil {
		return fmt.Errorf("parsing account balance: %w", err)
	}

	b := entities.Balance{
		AccountID: acc.ID,
		Balance:   balance,
		VegaTime:  vegaTime}

	err = t.balances.Add(b)
	if err != nil {
		return fmt.Errorf("adding balance to store: %w", err)
	}
	return nil
}

func (t *TransferResponse) addLedgerEntry(vle *vega.LedgerEntry, vegaTime time.Time) error {
	accFrom, err := t.obtainAccountWithID(vle.FromAccount, vegaTime)
	if err != nil {
		return fmt.Errorf("obtaining 'from' account: %w", err)
	}

	accTo, err := t.obtainAccountWithID(vle.ToAccount, vegaTime)
	if err != nil {
		return fmt.Errorf("obtaining 'to' account: %w", err)
	}

	quantity, err := decimal.NewFromString(vle.Amount)
	if err != nil {
		return fmt.Errorf("parsing amount string: %w", err)
	}

	le := entities.LedgerEntry{
		AccountFromID: accTo.ID,
		AccountToID:   accFrom.ID,
		Quantity:      quantity,
		VegaTime:      vegaTime,
		TransferTime:  time.Unix(0, vle.Timestamp),
		Reference:     vle.Reference,
		Type:          vle.Type,
	}

	err = t.ledger.Add(&le)
	if err != nil {
		return fmt.Errorf("adding to store: %w", err)
	}
	return nil
}

// Parse the vega account ID; if that account already exists in the db, fetch it; else create it.
func (t *TransferResponse) obtainAccountWithID(id string, vegaTime time.Time) (entities.Account, error) {
	a, err := entities.AccountFromAccountID(id)
	if err != nil {
		return entities.Account{}, fmt.Errorf("parsing account id: %w", err)
	}
	a.VegaTime = vegaTime
	err = t.accounts.Obtain(&a)
	if err != nil {
		return entities.Account{}, fmt.Errorf("obtaining account: %w", err)
	}
	return a, nil
}

func (t *TransferResponse) obtainAccountWithProto(va *vega.Account, vegaTime time.Time) (entities.Account, error) {
	a, err := entities.AccountFromProto(*va)
	if err != nil {
		return entities.Account{}, fmt.Errorf("obtaining account for balance: %w", err)
	}

	a.VegaTime = vegaTime
	err = t.accounts.Obtain(&a)
	if err != nil {
		return entities.Account{}, fmt.Errorf("obtaining account: %w", err)
	}
	return a, nil
}
