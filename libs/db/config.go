// Package db contain low db instance wrappers and helpers
package db

import (
	"fmt"
	"time"
)

type SelectorConfig map[string]*Config

type Config struct {
	Host                 string            `json:"host" yaml:"host"`
	Port                 int               `json:"port" yaml:"port"`
	Driver               string            `json:"driver" yaml:"driver"`
	Ssl                  bool              `json:"ssl" yaml:"ssl"`
	PreferSimpleProtocol bool              `json:"prefer_simple_protocol" yaml:"prefer_simple_protocol"`
	AdditionalParams     map[string]string `json:"additional_params" yaml:"additional_params"`

	Database string `json:"database" yaml:"database"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`

	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`
	MaxConnLifetime time.Duration `json:"max_conn_lifetime" yaml:"max_conn_lifetime"`

	DesiredVersion int    `json:"desired_version" yaml:"desired_version"`
	SQLDir         string `json:"sql_dir" yaml:"sql_dir"`
}

func (c *Config) FormatDriver() string {
	if c.Driver == "clickhouse" {
		return c.URLFormat()
	}
	return c.DSNFormat()
}

func (c *Config) DSNFormat() string {
	var sslmod string
	if c.Ssl {
		sslmod = "enable"
	} else {
		sslmod = "disable"
	}
	params := fmt.Sprintf(`host=%s port=%d dbname=%s user=%s password='%s' sslmode=%s`,
		c.Host,
		c.Port,
		c.Database,
		c.Username,
		c.Password,
		sslmod)
	return params
}

func (c *Config) URLFormat() string {
	params := fmt.Sprintf(`tcp://%s:%d`,
		c.Host,
		c.Port)
	return params
}
