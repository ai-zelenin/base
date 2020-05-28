package db

import (
	"database/sql"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"

	"git.pnhub.ru/core/libs/log"
)

func NewPGXStdlib(logger log.Logger, cfg *Config) (*sql.DB, error) {
	pgxStdlibCfg, err := pgx.ParseConfig(cfg.FormatDriver())
	if err != nil {
		return nil, err
	}

	pgxStdlibCfg.Logger = &PGXLogger{logger: log.ForkLogger(logger, cfg.Database)}
	pgxStdlibCfg.LogLevel = pgx.LogLevelError
	pgxStdlibCfg.PreferSimpleProtocol = cfg.PreferSimpleProtocol

	db := stdlib.OpenDB(*pgxStdlibCfg)
	if cfg.MaxConnLifetime > 0 {
		db.SetConnMaxLifetime(cfg.MaxConnLifetime)
	}
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	return db, nil
}
