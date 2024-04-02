package pgtest

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type ResetTestDBOp interface {
	run(ctx context.Context, q querier) error
	isResetTestDBOP()
}

func runResetTestDBOp(ctx context.Context, op ResetTestDBOp, testDB TestDB) error {
	conn, err := pgx.Connect(ctx, testDB.DataSourceName())
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := op.run(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

type resetTestDBDropAllTables struct {
	exclude []string
}

func (op *resetTestDBDropAllTables) isResetTestDBOP() {}
func (op *resetTestDBDropAllTables) run(ctx context.Context, q querier) error {
	return dropAllTables(ctx, q, &dropAllTablesArgs{
		exclude: op.exclude,
	})
}

func DropAllTablesExcept(names ...string) ResetTestDBOp {
	return &resetTestDBDropAllTables{
		exclude: names,
	}
}

func DropAllTables() ResetTestDBOp {
	return DropAllTablesExcept()
}

type resetTestDBTruncateAllTables struct {
	exclude []string
}

func (op *resetTestDBTruncateAllTables) isResetTestDBOP() {}
func (op *resetTestDBTruncateAllTables) run(ctx context.Context, q querier) error {
	return truncateAllTables(ctx, q, &truncateAllTablesArgs{
		exclude: op.exclude,
	})
}

func TruncateAllTablesExcept(names ...string) ResetTestDBOp {
	return &resetTestDBTruncateAllTables{
		exclude: names,
	}
}

func TruncateAllTables() ResetTestDBOp {
	return TruncateAllTablesExcept()
}
