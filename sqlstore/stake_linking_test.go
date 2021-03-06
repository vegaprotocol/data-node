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

package sqlstore_test

import (
	"context"
	"testing"
	"time"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/sqlstore"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	"code.vegaprotocol.io/vega/types/num"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStakeLinkingStore(t *testing.T) {
	t.Run("Upsert should add a stake linking record if it doesn't exist in the current block", testUpsertShouldAddNewInBlock)
	t.Run("Upsert should update a stake linking record if it already exists in the current block", testUpsertShouldUpdateExistingInBlock)
	t.Run("GetStake should return the most current version of each stake linking record and calculate the total stake available", testGetStake)
}

func setupStakeLinkingTest(t *testing.T, ctx context.Context) (*sqlstore.Blocks, *sqlstore.StakeLinking, *pgx.Conn) {
	t.Helper()
	DeleteEverything()

	bs := sqlstore.NewBlocks(connectionSource)
	sl := sqlstore.NewStakeLinking(connectionSource)

	config := NewTestConfig(testDBPort)
	conn, err := pgx.Connect(ctx, config.ConnectionConfig.GetConnectionString())
	require.NoError(t, err)

	return bs, sl, conn
}

func testUpsertShouldAddNewInBlock(t *testing.T) {
	testTimeout := time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	bs, sl, conn := setupStakeLinkingTest(t, ctx)

	var rowCount int
	assert.NoError(t, conn.QueryRow(ctx, "select count(*) from stake_linking").Scan(&rowCount))
	assert.Equal(t, 0, rowCount)

	block := addTestBlock(t, bs)
	stakingProtos := getStakingProtos()

	proto := stakingProtos[0]
	data, err := entities.StakeLinkingFromProto(proto, block.VegaTime)
	require.NoError(t, err)
	assert.NoError(t, sl.Upsert(context.Background(), data))

	assert.NoError(t, conn.QueryRow(ctx, "select count(*) from stake_linking").Scan(&rowCount))
	assert.Equal(t, 1, rowCount)
}

func testUpsertShouldUpdateExistingInBlock(t *testing.T) {
	testTimeout := time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	bs, sl, conn := setupStakeLinkingTest(t, ctx)

	var rowCount int
	assert.NoError(t, conn.QueryRow(ctx, "select count(*) from stake_linking").Scan(&rowCount))
	assert.Equal(t, 0, rowCount)

	block := addTestBlock(t, bs)
	stakingProtos := getStakingProtos()

	for _, proto := range stakingProtos {
		data, err := entities.StakeLinkingFromProto(proto, block.VegaTime)
		require.NoError(t, err)
		assert.NoError(t, sl.Upsert(context.Background(), data))
	}

	assert.NoError(t, conn.QueryRow(ctx, "select count(*) from stake_linking").Scan(&rowCount))
	assert.Equal(t, 2, rowCount)
}

func testGetStake(t *testing.T) {
	testTimeout := time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	bs, sl, conn := setupStakeLinkingTest(t, ctx)

	var rowCount int
	assert.NoError(t, conn.QueryRow(ctx, "select count(*) from stake_linking").Scan(&rowCount))
	assert.Equal(t, 0, rowCount)

	block := addTestBlock(t, bs)
	stakingProtos := getStakingProtos()

	for _, proto := range stakingProtos {
		data, err := entities.StakeLinkingFromProto(proto, block.VegaTime)
		require.NoError(t, err)
		assert.NoError(t, sl.Upsert(context.Background(), data))
	}

	assert.NoError(t, conn.QueryRow(ctx, "select count(*) from stake_linking").Scan(&rowCount))
	assert.Equal(t, 2, rowCount)

	partyID := entities.NewPartyID("cafed00d")

	currentBalance, links := sl.GetStake(ctx, partyID, entities.OffsetPagination{})
	want := num.NewUint(30002)
	assert.True(t, want.EQ(currentBalance))
	assert.Equal(t, 2, len(links))
}

func getStakingProtos() []*eventspb.StakeLinking {
	return []*eventspb.StakeLinking{
		{
			Id:              "deadbeef",
			Type:            eventspb.StakeLinking_TYPE_LINK,
			Ts:              time.Now().Unix(),
			Party:           "cafed00d",
			Amount:          "10000",
			Status:          eventspb.StakeLinking_STATUS_ACCEPTED,
			FinalizedAt:     time.Now().UnixNano(),
			TxHash:          "0xfe179560b9d0cc44c5fea54c2167c1cee7ccfcabf294752a4f43fb64ddffda85",
			BlockHeight:     1000000,
			BlockTime:       0,
			LogIndex:        100000,
			EthereumAddress: "TEST",
		},
		{
			Id:              "deadbeef",
			Type:            eventspb.StakeLinking_TYPE_LINK,
			Ts:              time.Now().Unix(),
			Party:           "cafed00d",
			Amount:          "10001",
			Status:          eventspb.StakeLinking_STATUS_ACCEPTED,
			FinalizedAt:     time.Now().UnixNano(),
			TxHash:          "0xfe179560b9d0cc44c5fea54c2167c1cee7ccfcabf294752a4f43fb64ddffda85",
			BlockHeight:     1000000,
			BlockTime:       0,
			LogIndex:        100000,
			EthereumAddress: "TEST",
		},
		{
			Id:              "deadbaad",
			Type:            eventspb.StakeLinking_TYPE_LINK,
			Ts:              time.Now().Unix(),
			Party:           "cafed00d",
			Amount:          "20001",
			Status:          eventspb.StakeLinking_STATUS_ACCEPTED,
			FinalizedAt:     time.Now().UnixNano(),
			TxHash:          "0xfe179560b9d0cc44c5fea54c2167c1cee7ccfcabf294752a4f43fb64ddffda85",
			BlockHeight:     1000000,
			BlockTime:       0,
			LogIndex:        100000,
			EthereumAddress: "TEST",
		},
	}
}
