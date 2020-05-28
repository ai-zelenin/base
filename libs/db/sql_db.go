package db

import (
	"database/sql"
)

func NewSQLDB(c *Config) (*sql.DB, error) {
	sqlDB, err := sql.Open(c.Driver, c.FormatDriver())
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(c.MaxConnLifetime)
	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}
	return sqlDB, nil
}
