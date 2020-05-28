package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/ClickHouse/clickhouse-go" // add clickhouse driver
	_ "github.com/jackc/pgx/v4/stdlib"      // add pgx driver
	_ "github.com/lib/pq"                   // add pq driver

	"go.uber.org/fx"

	"git.pnhub.ru/core/libs/log"
)

const DefaultDBKey = "default"

type Selector struct {
	ctx      context.Context
	logger   log.Logger
	cfgMap   SelectorConfig
	mx       sync.RWMutex
	pgxMap   map[string]*PGXNative
	sqlDBMap map[string]*sql.DB
}

func NewSelector(ctx context.Context, logger log.Logger, cfg SelectorConfig, lc fx.Lifecycle) (*Selector, error) {
	if cfg == nil || len(cfg) == 0 {
		return nil, fmt.Errorf("no cfg for db selecor")
	}
	dbs := &Selector{
		ctx:      ctx,
		logger:   logger,
		cfgMap:   cfg,
		pgxMap:   make(map[string]*PGXNative, len(cfg)),
		sqlDBMap: make(map[string]*sql.DB, len(cfg)),
	}

	for key, dbConfig := range cfg {
		err := NewMigrate(logger, dbConfig).Act()
		if err != nil {
			return nil, err
		}
		err = dbs.SetupDB(key, dbConfig)
		if err != nil {
			return nil, err
		}
	}

	if lc != nil {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return nil
			},
			OnStop: func(ctx context.Context) error {
				dbs.logger.Info("DB pools are closing")
				dbs.Close()
				dbs.logger.Info("DB pools has been closed")
				return nil
			},
		})
	}
	return dbs, nil
}

func (d *Selector) SetupDB(key string, cfg *Config) error {
	var db interface{}
	var err error
	switch cfg.Driver {
	case "pgx":
		db, err = NewPGXStdlib(d.logger, cfg)
		if err != nil {
			return err
		}
	case "pgx-native":
		db, err = NewPGXNative(d.ctx, d.logger, cfg)
		if err != nil {
			return err
		}
	default:
		db, err = NewSQLDB(cfg)
		if err != nil {
			return err
		}
	}
	d.setDB(key, cfg, db)
	return nil
}

func (d *Selector) setDB(key string, cfg *Config, i interface{}) {
	d.mx.Lock()
	defer d.mx.Unlock()
	switch db := i.(type) {
	case *PGXNative:
		d.pgxMap[key] = db
		d.cfgMap[key] = cfg
	case *sql.DB:
		d.sqlDBMap[key] = db
		d.cfgMap[key] = cfg
	}
}

func (d *Selector) Cfg(keys ...string) *Config {
	d.mx.RLock()
	defer d.mx.RUnlock()
	if len(keys) == 0 {
		return d.cfgMap[DefaultDBKey]
	}
	key := keys[0]
	cfg, ok := d.cfgMap[key]
	if !ok {
		d.logger.Fatal("no db with key %s", key)
	}
	return cfg
}

func (d *Selector) DB(keys ...string) *sql.DB {
	d.mx.RLock()
	defer d.mx.RUnlock()
	if len(keys) == 0 {
		return d.sqlDBMap[DefaultDBKey]
	}
	key := keys[0]
	db, ok := d.sqlDBMap[key]
	if !ok {
		d.logger.Fatal("no db with key %s", key)
	}
	return db
}

func (d *Selector) PGXNative(keys ...string) *PGXNative {
	d.mx.RLock()
	defer d.mx.RUnlock()
	if len(keys) == 0 {
		return d.pgxMap[DefaultDBKey]
	}
	key := keys[0]
	db, ok := d.pgxMap[key]
	if !ok {
		d.logger.Fatal("no db with key %s", key)
	}
	return db
}

func (d *Selector) Close() {
	d.mx.Lock()
	defer d.mx.Unlock()

	for k, db := range d.sqlDBMap {
		d.logger.Debugf("closing sql.DB %s", k)
		err := db.Close()
		if err != nil {
			d.logger.Error(err)
		}
	}
	for k, pgx := range d.pgxMap {
		d.logger.Infof("closing pgx %s", k)
		pgx.Close()
	}
}
