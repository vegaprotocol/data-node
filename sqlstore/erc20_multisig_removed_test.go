package sqlstore_test

import (
	"context"
	"testing"
	"time"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/sqlstore"
	eventspb "code.vegaprotocol.io/protos/vega/events/v1"
	vgcrypto "code.vegaprotocol.io/shared/libs/crypto"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestERC20MultiSigRemoved(t *testing.T) {
	t.Run("Adding a single bundle", testAddRemovedSigner)
	t.Run("Get with filters", testGetRemovedWithFilters)
}

func setupERC20MultiSigRemovedStoreTests(t *testing.T, ctx context.Context) (*sqlstore.ERC20MultiSigSignerRemoved, *pgx.Conn) {
	t.Helper()
	err := testStore.DeleteEverything()
	require.NoError(t, err)
	ms := sqlstore.NewERC20MultiSigSignerRemoved(testStore)

	config := NewTestConfig(testDBPort)

	conn, err := pgx.Connect(ctx, connectionString(config))
	require.NoError(t, err)

	return ms, conn
}

func testAddRemovedSigner(t *testing.T) {
	testTimeout := time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	ms, conn := setupERC20MultiSigRemovedStoreTests(t, ctx)

	var rowCount int

	err := conn.QueryRow(ctx, `select count(*) from erc20_multisig_signer_removed`).Scan(&rowCount)
	require.NoError(t, err)
	assert.Equal(t, 0, rowCount)

	sa := getTestSignerRemoved(t, "fc677151d0c93726", vgcrypto.RandomHash(), "12")
	err = ms.Add(sa)
	require.NoError(t, err)

	err = conn.QueryRow(ctx, `select count(*) from erc20_multisig_signer_removed`).Scan(&rowCount)
	require.NoError(t, err)
	assert.Equal(t, 1, rowCount)

	// now add a duplicate and check we still remain with one
	err = ms.Add(sa)
	require.NoError(t, err)

	err = conn.QueryRow(ctx, `select count(*) from erc20_multisig_signer_removed`).Scan(&rowCount)
	require.NoError(t, err)
	assert.Equal(t, 1, rowCount)
}

func testGetRemovedWithFilters(t *testing.T) {
	testTimeout := time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	ms, conn := setupERC20MultiSigRemovedStoreTests(t, ctx)

	var rowCount int
	vID1 := "fc677151d0c93726"
	vID2 := "15d1d5fefa8988eb"
	submitter1 := "3451cf3d0b743312"
	submitter2 := "5940a49413c52516"

	err := conn.QueryRow(ctx, `select count(*) from erc20_multisig_signer_added`).Scan(&rowCount)
	require.NoError(t, err)
	assert.Equal(t, 0, rowCount)

	err = ms.Add(getTestSignerRemoved(t, vID1, submitter1, "12"))
	require.NoError(t, err)
	err = ms.Add(getTestSignerRemoved(t, vID1, submitter2, "12"))
	require.NoError(t, err)

	// same validator and submitter different epoch
	err = ms.Add(getTestSignerRemoved(t, vID1, submitter1, "24"))
	require.NoError(t, err)

	// same epoch and submitter different validator
	err = ms.Add(getTestSignerRemoved(t, vID2, submitter1, "12"))
	require.NoError(t, err)

	res, err := ms.GetByValidatorID(ctx, vID1, submitter1, nil, entities.Pagination{})
	require.NoError(t, err)
	require.Len(t, res, 2)
	assert.Equal(t, vID1, res[0].ValidatorID.String())
	assert.Equal(t, submitter1, res[0].Submitter.String())
	assert.Equal(t, vID1, res[1].ValidatorID.String())
	assert.Equal(t, submitter1, res[1].Submitter.String())

	epochID := int64(24)
	res, err = ms.GetByValidatorID(ctx, vID1, submitter1, &epochID, entities.Pagination{})
	require.NoError(t, err)
	require.Len(t, res, 1)
	assert.Equal(t, vID1, res[0].ValidatorID.String())
	assert.Equal(t, submitter1, res[0].Submitter.String())
	assert.Equal(t, epochID, res[0].EpochID)
}

func getTestSignerRemoved(t *testing.T, validatorID, submitter, epochSeq string) *entities.ERC20MultiSigSignerRemoved {
	t.Helper()
	vgcrypto.RandomHash()
	ns, err := entities.ERC20MultiSigSignerRemovedFromProto(
		&eventspb.ERC20MultiSigSignerRemoved{
			SignatureSubmitters: []*eventspb.ERC20MulistSigSignerRemovedSubmitter{
				{
					SignatureId: vgcrypto.RandomHash(),
					Submitter:   submitter,
				},
			},
			ValidatorId: validatorID,
			OldSigner:   vgcrypto.RandomHash(),
			Nonce:       "nonce",
			EpochSeq:    epochSeq,
			Timestamp:   time.Unix(10000, 13).UnixNano(),
		},
	)
	require.NoError(t, err)
	require.Len(t, ns, 1)
	return ns[0]
}
