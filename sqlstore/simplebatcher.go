package sqlstore

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

type SimpleBatcher struct {
	pending     [][]interface{}
	tableName   string
	columnNames []string
}

func NewSimpleBatcher(tableName string, columnNames []string) SimpleBatcher {
	return SimpleBatcher{
		tableName:   tableName,
		columnNames: columnNames,
		pending:     make([][]interface{}, 0, 1000),
	}
}

type simpleEntity interface {
	ToRow() []interface{}
}

func (b *SimpleBatcher) Add(entity simpleEntity) {
	row := entity.ToRow()
	b.pending = append(b.pending, row)
}

func (b *SimpleBatcher) Flush(ctx context.Context, pool TxDelegator) error {
	copyCount, err := pool.CopyFrom(
		ctx,
		pgx.Identifier{b.tableName},
		b.columnNames,
		pgx.CopyFromRows(b.pending),
	)
	if err != nil {
		return fmt.Errorf("failed to copy %s entries into database:%w", b.tableName, err)
	}

	if copyCount != int64(len(b.pending)) {
		return fmt.Errorf("copied %d %s rows into the database, expected to copy %d",
			copyCount,
			b.tableName,
			len(b.pending))
	}

	b.pending = b.pending[:0]
	return nil
}
