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

func TestERC20MultiSigAdded(t *testing.T) {
	t.Run("Adding a single bundle", testAddSigner)
	t.Run("Get with filters", testGetWithFilters)
}

func setupERC20MultiSigAddedStoreTests(t *testing.T, ctx context.Context) (*sqlstore.ERC20MultiSigSignerAdded, *pgx.Conn) {
	t.Helper()
	err := testStore.DeleteEverything()
	require.NoError(t, err)
	ms := sqlstore.NewERC20MultiSigSignerAdded(testStore)

	config := NewTestConfig(testDBPort)

	conn, err := pgx.Connect(ctx, connectionString(config))
	require.NoError(t, err)

	return ms, conn
}

func testAddSigner(t *testing.T) {
	testTimeout := time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	ms, conn := setupERC20MultiSigAddedStoreTests(t, ctx)

	var rowCount int

	err := conn.QueryRow(ctx, `select count(*) from erc20_multisig_signer_added`).Scan(&rowCount)
	require.NoError(t, err)
	assert.Equal(t, 0, rowCount)

	sa := getTestSignerAdded(t, "fc677151d0c93726", "12")
	err = ms.Add(sa)
	require.NoError(t, err)

	err = conn.QueryRow(ctx, `select count(*) from erc20_multisig_signer_added`).Scan(&rowCount)
	require.NoError(t, err)
	assert.Equal(t, 1, rowCount)

	// now add a duplicate and check we still remain with one
	err = ms.Add(sa)
	require.NoError(t, err)

	err = conn.QueryRow(ctx, `select count(*) from erc20_multisig_signer_added`).Scan(&rowCount)
	require.NoError(t, err)
	assert.Equal(t, 1, rowCount)
}

func testGetWithFilters(t *testing.T) {
	testTimeout := time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	ms, conn := setupERC20MultiSigAddedStoreTests(t, ctx)

	var rowCount int
	vID1 := "fc677151d0c93726"
	vID2 := "15d1d5fefa8988eb"

	err := conn.QueryRow(ctx, `select count(*) from erc20_multisig_signer_added`).Scan(&rowCount)
	require.NoError(t, err)
	assert.Equal(t, 0, rowCount)

	err = ms.Add(getTestSignerAdded(t, vID1, "12"))
	require.NoError(t, err)

	// same validator different epoch
	err = ms.Add(getTestSignerAdded(t, vID1, "24"))
	require.NoError(t, err)

	// same epoch different validator
	err = ms.Add(getTestSignerAdded(t, vID2, "12"))
	require.NoError(t, err)

	res, err := ms.GetByValidatorID(ctx, vID1, nil, entities.Pagination{})
	require.NoError(t, err)
	require.Len(t, res, 2)
	assert.Equal(t, vID1, res[0].ValidatorID.String())
	assert.Equal(t, vID1, res[1].ValidatorID.String())

	res, err = ms.GetByEpochID(ctx, 12, entities.Pagination{})
	require.NoError(t, err)
	require.Len(t, res, 2)
	assert.Equal(t, int64(12), res[0].EpochID)
	assert.Equal(t, int64(12), res[1].EpochID)

	epoch := int64(12)
	res, err = ms.GetByValidatorID(ctx, vID1, &epoch, entities.Pagination{})
	require.NoError(t, err)
	require.Len(t, res, 1)
	assert.Equal(t, vID1, res[0].ValidatorID.String())
	assert.Equal(t, int64(12), res[0].EpochID)
}

func getTestSignerAdded(t *testing.T, validatorID string, epochSeq string) *entities.ERC20MultiSigSignerAdded {
	t.Helper()
	vgcrypto.RandomHash()
	ns, err := entities.ERC20MultiSigSignerAddedFromProto(
		&eventspb.ERC20MultiSigSignerAdded{
			SignatureId: vgcrypto.RandomHash(),
			ValidatorId: validatorID,
			NewSigner:   vgcrypto.RandomHash(),
			Submitter:   vgcrypto.RandomHash(),
			Nonce:       "nonce",
			EpochSeq:    epochSeq,
			Timestamp:   time.Unix(10000, 13).UnixNano(),
		},
	)
	require.NoError(t, err)
	return ns
}
