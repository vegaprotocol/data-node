package sqlstore

import (
	"context"

	"code.vegaprotocol.io/data-node/entities"
	"github.com/georgysavva/scany/pgxscan"
)

type Blocks struct {
	*SqlStore
}

func NewBlocks(sqlStore *SqlStore) *Blocks {
	b := &Blocks{
		SqlStore: sqlStore,
	}
	return b
}

func (bs *Blocks) Add(b entities.Block) error {
	ctx := context.Background()
	_, err := bs.pool.Exec(ctx,
		`insert into blocks(vega_time, height, hash) values ($1, $2, $3)`,
		b.VegaTime, b.Height, b.Hash)
	return err
}

func (bs *Blocks) GetAll() ([]entities.Block, error) {
	ctx := context.Background()
	blocks := []entities.Block{}
	err := pgxscan.Select(ctx, bs.pool, &blocks,
		`SELECT vega_time, height, hash
		FROM blocks
		ORDER BY vega_time desc`)
	return blocks, err
}
