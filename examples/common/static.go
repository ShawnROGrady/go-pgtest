package common

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	pgxdb "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*
var migrationFiles embed.FS

type MigrationTarget interface {
	isMigrationTarget()
	migrate(sourceName string, sourceInstance source.Driver) (*migrate.Migrate, error)
}

type targetDB struct {
	inst database.Driver
	name string
}

func (*targetDB) isMigrationTarget() {}
func (t *targetDB) migrate(sourceName string, sourceInstance source.Driver) (*migrate.Migrate, error) {
	dbName := t.name
	if dbName == "" {
		dbName = "postgres"
	}

	return migrate.NewWithInstance(sourceName, sourceInstance, dbName, t.inst)
}

type targetURL struct {
	url string
}

func (*targetURL) isMigrationTarget() {}
func (t *targetURL) migrate(sourceName string, sourceInstance source.Driver) (*migrate.Migrate, error) {
	return migrate.NewWithSourceInstance(sourceName, sourceInstance, t.url)
}

type MigrationsSource struct {
	src source.Driver
}

func Migrations() MigrationsSource {
	src, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		panic(fmt.Sprintf("load migration files: %s", err))
	}

	return MigrationsSource{src: src}
}

func (m MigrationsSource) ForTarget(target MigrationTarget) (*migrate.Migrate, error) {
	return target.migrate("embed", m.src)
}

func (m MigrationsSource) ForURL(target string) (*migrate.Migrate, error) {
	return m.ForTarget(&targetURL{url: target})
}

func (m MigrationsSource) ForInstance(target database.Driver) (*migrate.Migrate, error) {
	return m.ForTarget(&targetDB{inst: target, name: "postgres"})
}

func (m MigrationsSource) ForPgxPoolWithConf(target *pgxpool.Pool, conf *pgxdb.Config) (*migrate.Migrate, error) {
	stdDB := stdlib.OpenDBFromPool(target)
	dbInst, err := pgxdb.WithInstance(stdDB, conf)
	if err != nil {
		return nil, fmt.Errorf("load pgx db instance: %w", err)
	}

	return m.ForInstance(dbInst)
}

func (m MigrationsSource) ForPgxPool(target *pgxpool.Pool) (*migrate.Migrate, error) {
	return m.ForPgxPoolWithConf(target, &pgxdb.Config{})
}
