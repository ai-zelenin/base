package db

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"git.pnhub.ru/core/libs/log"
)

type PGXNative struct {
	*pgxpool.Pool
}

func NewPGXNative(ctx context.Context, logger log.Logger, cfg *Config) (*PGXNative, error) {
	dbPoolCfg, err := pgxpool.ParseConfig(cfg.FormatDriver())
	if err != nil {
		return nil, err
	}
	dbPoolCfg.ConnConfig.Logger = &PGXLogger{logger: log.ForkLogger(logger, cfg.Database)}
	dbPoolCfg.ConnConfig.LogLevel = pgx.LogLevelError

	if cfg.MaxConnLifetime > 0 {
		dbPoolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxOpenConns > 0 {
		dbPoolCfg.MaxConns = int32(cfg.MaxOpenConns)
	}

	pool, err := pgxpool.ConnectConfig(ctx, dbPoolCfg)
	if err != nil {
		return nil, err
	}
	out := &PGXNative{Pool: pool}
	err = out.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (p *PGXNative) Ping(ctx context.Context) error {
	c, err := p.Acquire(ctx)
	if err != nil {
		return err
	}
	defer c.Release()
	err = c.Conn().Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}
