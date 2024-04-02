package pgtest

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type querier interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
}

func dropDatabase(ctx context.Context, q querier, name string) error {
	query := fmt.Sprintf("DROP DATABASE %q;", name)
	if _, err := q.Exec(ctx, query); err != nil {
		return err
	}

	return nil
}

type pgTable struct {
	name  string
	owner string
}

func getAllTablesInCurrentSchema(ctx context.Context, q querier) ([]pgTable, error) {
	rows, err := q.Query(ctx, `SELECT tablename, tableowner FROM pg_tables WHERE schemaname = (SELECT current_schema());`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []pgTable
	for rows.Next() {
		var table pgTable
		if err := rows.Scan(&table.name, &table.owner); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

type truncateTableArgs struct {
	name            string
	restartIdentity bool
	cascade         bool
}

func (args *truncateTableArgs) query() string {
	var b strings.Builder
	b.WriteString("TRUNCATE ")
	b.WriteString(strconv.Quote(args.name))

	if args.restartIdentity {
		b.WriteString(" RESTART IDENTITY")
	}

	if args.cascade {
		b.WriteString(" CASCADE")
	}

	b.WriteRune(';')
	return b.String()
}

func truncateTable(ctx context.Context, q querier, args *truncateTableArgs) error {
	_, err := q.Exec(ctx, args.query())
	return err
}

type truncateAllTablesArgs struct {
	exclude []string
}

func (args *truncateAllTablesArgs) skip(table pgTable) bool {
	if args == nil {
		return false
	}

	if len(args.exclude) == 0 {
		return false
	}

	return slices.Contains(args.exclude, table.name)
}

func truncateAllTables(ctx context.Context, q querier, args *truncateAllTablesArgs) error {
	tables, err := getAllTablesInCurrentSchema(ctx, q)
	if err != nil {
		return fmt.Errorf("get tables in current schema: %w", err)
	}

	toTruncate := slices.DeleteFunc(tables, args.skip)
	for _, table := range toTruncate {
		if err := truncateTable(ctx, q, &truncateTableArgs{
			name:            table.name,
			restartIdentity: true,
			cascade:         true,
		}); err != nil {
			return fmt.Errorf("truncate %s: %w", table.name, err)
		}
	}

	return err
}

type dropTableArgs struct {
	name    string
	cascade bool
}

func (args *dropTableArgs) query() string {
	var b strings.Builder
	b.WriteString("DROP TABLE ")
	b.WriteString(strconv.Quote(args.name))

	if args.cascade {
		b.WriteString(" CASCADE")
	}

	b.WriteRune(';')
	return b.String()
}

func dropTable(ctx context.Context, q querier, args *dropTableArgs) error {
	_, err := q.Exec(ctx, args.query())
	return err
}

type dropAllTablesArgs struct {
	exclude []string
}

func (args *dropAllTablesArgs) skip(table pgTable) bool {
	if args == nil {
		return false
	}

	if len(args.exclude) == 0 {
		return false
	}

	return slices.Contains(args.exclude, table.name)
}

func dropAllTables(ctx context.Context, q querier, args *dropAllTablesArgs) error {
	tables, err := getAllTablesInCurrentSchema(ctx, q)
	if err != nil {
		return fmt.Errorf("get tables in current schema: %w", err)
	}

	toDrop := slices.DeleteFunc(tables, args.skip)
	for _, table := range toDrop {
		if err := dropTable(ctx, q, &dropTableArgs{
			name:    table.name,
			cascade: true,
		}); err != nil {
			return fmt.Errorf("drop %s: %w", table.name, err)
		}
	}

	return err
}

func getAllDatabases(ctx context.Context, q querier) ([]string, error) {
	rows, err := q.Query(ctx, `SELECT datname FROM pg_database;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		dbNames = append(dbNames, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dbNames, nil
}
