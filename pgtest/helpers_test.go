package pgtest

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/pashagolub/pgxmock/v3"
)

func newMockQuerier(t testing.TB) pgxmock.PgxCommonIface {
	t.Helper()
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("unexpected error creating mock pgx pool: %s", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

type truncateAllTablesTestCase struct {
	tablesInCurrentSchema [][]any
	excludeNames          []string
	expectTruncated       []string
}

func (tc *truncateAllTablesTestCase) setUpMock(t *testing.T) pgxmock.PgxCommonIface {
	t.Helper()
	mockQuerier := newMockQuerier(t)

	mockQuerier.
		ExpectQuery(regexp.QuoteMeta(
			`SELECT tablename, tableowner FROM pg_tables WHERE schemaname = (SELECT current_schema());`,
		)).
		WillReturnRows(
			pgxmock.NewRows([]string{"tablename", "tableowner"}).
				AddRows(tc.tablesInCurrentSchema...),
		).
		RowsWillBeClosed().
		Times(1)

	for _, table := range tc.expectTruncated {
		mockQuerier.
			ExpectExec(regexp.QuoteMeta(fmt.Sprintf(
				"TRUNCATE %q RESTART IDENTITY CASCADE", table,
			))).
			WillReturnResult(pgxmock.NewResult("TRUNCATE", 1)).
			Times(1)
	}

	return mockQuerier
}

func TestTruncateAllTablesSuccess(t *testing.T) {
	testCases := map[string]truncateAllTablesTestCase{
		"no_exclude": truncateAllTablesTestCase{
			tablesInCurrentSchema: [][]any{
				{"table1", "me"},
				{"table2", "me"},
				{"table3", "me"},
			},
			excludeNames:    nil,
			expectTruncated: []string{"table1", "table2", "table3"},
		},

		"exclude_contains_table_in_current_schema": truncateAllTablesTestCase{
			tablesInCurrentSchema: [][]any{
				{"table1", "me"},
				{"table2", "me"},
				{"table3", "me"},
			},
			excludeNames:    []string{"table2"},
			expectTruncated: []string{"table1", "table3"},
		},

		"exclude_contains_table_not_in_current_schema": truncateAllTablesTestCase{
			tablesInCurrentSchema: [][]any{
				{"table1", "me"},
				{"table2", "me"},
				{"table3", "me"},
			},
			excludeNames:    []string{"other_table"},
			expectTruncated: []string{"table1", "table2", "table3"},
		},
	}

	for testName, tc := range testCases {
		t.Run(testName, func(t *testing.T) {
			ctx := context.Background()
			mockQuerier := tc.setUpMock(t)

			if err := truncateAllTables(ctx, mockQuerier, &truncateAllTablesArgs{
				exclude: tc.excludeNames,
			}); err != nil {
				t.Errorf("unexpected error returned by truncateAllTables: %s", err)
			}

			if err := mockQuerier.ExpectationsWereMet(); err != nil {
				t.Errorf("mock querier has unfulfilled expectations: %s", err)
			}
		})
	}
}

type dropAllTablesTestCase struct {
	tablesInCurrentSchema [][]any
	excludeNames          []string
	expectDropped         []string
}

func (tc *dropAllTablesTestCase) setUpMock(t *testing.T) pgxmock.PgxCommonIface {
	t.Helper()
	mockQuerier := newMockQuerier(t)

	mockQuerier.
		ExpectQuery(regexp.QuoteMeta(
			`SELECT tablename, tableowner FROM pg_tables WHERE schemaname = (SELECT current_schema());`,
		)).
		WillReturnRows(
			pgxmock.NewRows([]string{"tablename", "tableowner"}).
				AddRows(tc.tablesInCurrentSchema...),
		).
		RowsWillBeClosed().
		Times(1)

	for _, table := range tc.expectDropped {
		mockQuerier.
			ExpectExec(regexp.QuoteMeta(fmt.Sprintf(
				"DROP TABLE %q CASCADE;", table,
			))).
			WillReturnResult(pgxmock.NewResult("TRUNCATE", 1)).
			Times(1)
	}

	return mockQuerier
}

func TestDropAllTablesSuccess(t *testing.T) {
	testCases := map[string]dropAllTablesTestCase{
		"no_exclude": dropAllTablesTestCase{
			tablesInCurrentSchema: [][]any{
				{"table1", "me"},
				{"table2", "me"},
				{"table3", "me"},
			},
			excludeNames:  nil,
			expectDropped: []string{"table1", "table2", "table3"},
		},

		"exclude_contains_table_in_current_schema": dropAllTablesTestCase{
			tablesInCurrentSchema: [][]any{
				{"table1", "me"},
				{"table2", "me"},
				{"table3", "me"},
			},
			excludeNames:  []string{"table2"},
			expectDropped: []string{"table1", "table3"},
		},

		"exclude_contains_table_not_in_current_schema": dropAllTablesTestCase{
			tablesInCurrentSchema: [][]any{
				{"table1", "me"},
				{"table2", "me"},
				{"table3", "me"},
			},
			excludeNames:  []string{"other_table"},
			expectDropped: []string{"table1", "table2", "table3"},
		},
	}

	for testName, tc := range testCases {
		t.Run(testName, func(t *testing.T) {
			ctx := context.Background()
			mockQuerier := tc.setUpMock(t)

			if err := dropAllTables(ctx, mockQuerier, &dropAllTablesArgs{
				exclude: tc.excludeNames,
			}); err != nil {
				t.Errorf("unexpected error returned by dropAllTables: %s", err)
			}

			if err := mockQuerier.ExpectationsWereMet(); err != nil {
				t.Errorf("mock querier has unfulfilled expectations: %s", err)
			}
		})
	}
}
