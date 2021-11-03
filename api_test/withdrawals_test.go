package api_test

import (
	"context"
	"testing"
	"time"

	apipb "code.vegaprotocol.io/protos/data-node/api/v1"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vega/events"
	"code.vegaprotocol.io/vega/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithdrawals(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimout)
	defer cancel()

	ts := NewTestServer(t, ctx, true)
	defer ts.Close()

	PublishEvents(t, ctx, ts.broker, func(be *eventspb.BusEvent) (events.Event, error) {
		withdrawal := be.GetWithdrawal()
		require.NotNil(t, withdrawal)
		e := events.NewWithdrawalEvent(ctx, types.Withdrawal{
			ID:             withdrawal.Id,
			PartyID:        withdrawal.PartyId,
			Amount:         mustUintFromString(withdrawal.Amount),
			Asset:          withdrawal.Asset,
			Status:         withdrawal.Status,
			Ref:            withdrawal.Ref,
			ExpirationDate: withdrawal.Expiry,
			TxHash:         withdrawal.TxHash,
			CreationDate:   withdrawal.CreatedTimestamp,
			WithdrawalDate: withdrawal.WithdrawnTimestamp,
		})
		return e, nil
	}, "withdrawals-events.golden")

	client := apipb.NewTradingDataServiceClient(ts.conn)
	require.NotNil(t, client)

	withdrawalID := "af6e66ee1e1a643338f55b8dfe00129b09b926a997edddf1f10e76b31c65cdad"
	withdrawalPartyID := "c5fdc709b3464ca10292437ce493dc0e497b2c3ea22a5fde714c4e487b93011d"

	var resp *apipb.WithdrawalsResponse
	var err error

loop:
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("test timeout")
		case <-time.Tick(50 * time.Millisecond):
			resp, err = client.Withdrawals(ctx, &apipb.WithdrawalsRequest{
				PartyId: withdrawalPartyID,
			})
			require.NotNil(t, resp)
			require.NoError(t, err)
			if len(resp.Withdrawals) > 0 {
				break loop
			}
		}
	}

	assert.NoError(t, err)
	assert.Equal(t, withdrawalID, resp.Withdrawals[0].Id)
	assert.Equal(t, withdrawalPartyID, resp.Withdrawals[0].PartyId)
}
