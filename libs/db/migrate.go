package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/clickhouse"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"

	"git.pnhub.ru/core/libs/log"
)

type Migrate struct {
	cfg    *Config
	logger log.Logger
}

func NewMigrate(logger log.Logger, cfg *Config) *Migrate {
	return &Migrate{cfg: cfg, logger: log.ForkLogger(logger)}
}

func (m *Migrate) Act() error {
	if m.cfg.DesiredVersion > 0 && m.cfg.SQLDir != "" {
		driver, err := m.CreateMigrationDriver()
		if err != nil {
			return err
		}
		sqlDir := fmt.Sprintf("file://%s", m.cfg.SQLDir)
		fd, err := new(file.File).Open(sqlDir)
		if err != nil {
			return err
		}
		version, dirty, err := driver.Version()
		if err != nil {
			return err
		}
		m.logger.Infof("current DB migration: %s - version:%d dirty:%t", m.cfg.Database, version, dirty)
		if m.cfg.DesiredVersion != version {
			instance, err := migrate.NewWithInstance("file", fd, m.cfg.Database, driver)
			if err != nil {
				return err
			}
			err = instance.Migrate(uint(m.cfg.DesiredVersion))
			if err != nil {
				return err
			}
			version, dirty, err := instance.Version()
			if err != nil {
				return err
			}
			m.logger.Infof("migrated DB: %s - version:%d dirty:%t", m.cfg.Database, version, dirty)
		}
	}
	return nil
}

func (m *Migrate) CreateMigrationDriver() (database.Driver, error) {
	switch m.cfg.Driver {
	case "postgres", "pgx", "pgx-native":
		db, err := sql.Open("pgx", m.cfg.FormatDriver())
		if err != nil {
			return nil, err
		}
		return postgres.WithInstance(db, &postgres.Config{})
	case "clickhouse":
		db, err := sql.Open(m.cfg.Driver, m.cfg.FormatDriver())
		if err != nil {
			return nil, err
		}
		return clickhouse.WithInstance(db, &clickhouse.Config{})
	default:
		return nil, fmt.Errorf("unknown driver for migrations tool")
	}
}
