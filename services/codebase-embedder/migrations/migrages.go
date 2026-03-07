package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/zgsm-ai/codebase-indexer/internal/config"
)

//go:embed sql/*.sql
var migrateFS embed.FS

// migrate create -dir migrations/sql -ext sql  init

const datasourcePostgres = "postgres"

func AutoMigrate(c config.Database) error {
	if !c.AutoMigrate.Enable {
		fmt.Println("AutoMigrate is disabled.")
		return nil
	}
	fmt.Println("===start to migrate===")
	db, err := sql.Open(datasourcePostgres, c.DataSource)
	if err != nil {
		return err
	}

	d, err := iofs.New(migrateFS, "sql")
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		d,
		datasourcePostgres, driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	fmt.Println("===auto migrate successfully===")
	return nil
}
