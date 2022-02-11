package sqlstore_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/data-node/sqlstore"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

func Test_MarketData_Add(t *testing.T) {
	t.Run("Add should insert a valid market data record", shouldInsertAValidMarketDataRecord)
	t.Run("Add should return an error if the vega block does not exist", shouldErrorIfNoVegaBlock)
}

func shouldInsertAValidMarketDataRecord(t *testing.T) {
	bs := sqlstore.NewBlocks(testStore)
	md := sqlstore.NewMarketData(testStore)

	defer testStore.Stop()

	err := testStore.DeleteEverything()
	assert.NoError(t, err)

	config := sqlstore.NewDefaultConfig()
	config.Port = 15432

	connStr := connectionString(config)

	testTimeout := time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, connStr)
	assert.NoError(t, err)
	var rowCount int

	err = conn.QueryRow(ctx, `select count(*) from market_data`).Scan(&rowCount)
	assert.NoError(t, err)
	assert.Equal(t, 0, rowCount)

	block := addTestBlock(t, bs)

	err = md.Add(ctx, &entities.MarketData{
		VegaTime: block.VegaTime,
	})
	assert.NoError(t, err)

	err = conn.QueryRow(ctx, `select count(*) from market_data`).Scan(&rowCount)
	assert.NoError(t, err)
	assert.Equal(t, 1, rowCount)
}

func shouldErrorIfNoVegaBlock(t *testing.T) {
	md := sqlstore.NewMarketData(testStore)

	defer testStore.Stop()

	err := testStore.DeleteEverything()
	assert.NoError(t, err)

	config := sqlstore.NewDefaultConfig()
	config.Port = 15432

	connStr := connectionString(config)

	testTimeout := time.Second * 10
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, connStr)
	assert.NoError(t, err)
	var rowCount int

	err = conn.QueryRow(ctx, `select count(*) from market_data`).Scan(&rowCount)
	assert.NoError(t, err)
	assert.Equal(t, 0, rowCount)

	err = md.Add(ctx, &entities.MarketData{
		VegaTime: time.Now().Truncate(time.Microsecond),
	})
	assert.Error(t, err)

	err = conn.QueryRow(ctx, `select count(*) from market_data`).Scan(&rowCount)
	assert.NoError(t, err)
	assert.Equal(t, 0, rowCount)
}

func connectionString(config sqlstore.Config) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database)
}
