package pharvester

import (
	"time"

	"git.pnhub.ru/core/libs/base"
)

func NewConfig() (*Config, error) {
	cfg := new(Config)
	err := base.ReadConfig(cfg, base.ReadCfgPathFlag())
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

type Config struct {
	ImportFilePath  string           `json:"import_file_path" yaml:"import_file_path"`
	ValidatorConfig *ValidatorConfig `json:"validator_config" yaml:"validator_config"`
	SourceConfigs   []*SourceConfig  `json:"source_configs" yaml:"source_configs"`
}

type ValidatorConfig struct {
	Host             string        `json:"host" yaml:"host"`
	Path             string        `json:"path" yanl:"path"`
	Timeout          time.Duration `json:"timeout" yaml:"timeout"`
	NumberOfRequests int           `json:"number_of_requests" yaml:"number_of_requests"`
	Threads          int           `json:"threads" yaml:"threads"`
}

type SourceConfig struct {
	Skip        bool          `json:"skip" yanl:"skip"`
	Delay       time.Duration `json:"delay" yaml:"delay"`
	RandomDelay time.Duration `json:"random_delay" yaml:"random_delay"`
	Threads     int           `json:"threads" yaml:"threads"`

	URL          string      `json:"url" yaml:"url"`
	Depth        int         `json:"depth" yaml:"depth"`
	FollowRegexp string      `json:"follow_regexp" yaml:"follow_regexp"`
	Selectors    []*Selector `json:"selectors" yaml:"selectors"`
	Dir          string      `json:"dir" yaml:"dir"`
}

type Selector struct {
	Target           string            `json:"target" yaml:"target"`
	EnableValidation bool              `json:"enable_validation" yaml:"enable_validation"`
	Selector         string            `json:"selector" yaml:"selector"`
	FilterRegexp     string            `json:"filter_regexp" yaml:"filter_regexp"`
	Mapping          map[string]string `json:"mapping" yaml:"mapping"`
	Array            []string          `json:"array" yaml:"array"`
}
