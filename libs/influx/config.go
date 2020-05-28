package influx

import (
	"time"
)

type SelectorConfig map[string]*Config

type Config struct {
	URL       string `json:"url" yaml:"url"`
	Username  string `json:"username" yaml:"username"`
	Password  string `json:"password" yaml:"password"`
	UserAgent string `json:"user_agent" yaml:"user_agent"`
	Database  string `json:"database" yaml:"database"`

	RetentionPolicy       string        `json:"retention_policy" yaml:"retention_policy"`
	Consistency           string        `json:"consistency" yaml:"consistency"`
	ContentEncoding       string        `json:"content_encoding" yaml:"content_encoding"`
	Precision             string        `json:"precision" yaml:"precision"`
	Timeout               time.Duration `json:"timeout" yaml:"timeout"`
	ResponseHeaderTimeout time.Duration `json:"response_header_timeout" yaml:"response_header_timeout"`
	IdleConnTimeout       time.Duration `json:"idle_conn_timeout" yaml:"idle_conn_timeout"`
}
