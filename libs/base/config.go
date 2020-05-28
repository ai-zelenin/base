package base

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"gopkg.in/yaml.v2"

	"git.pnhub.ru/core/libs/db"
	"git.pnhub.ru/core/libs/influx"
	"git.pnhub.ru/core/libs/kfk"
	"git.pnhub.ru/core/libs/log"
)

type Config struct {
	fx.Out

	Zap *log.ZapConfig `json:"log" yaml:"log"`

	HTTPServer *HTTPServerConfig `json:"http_server" yaml:"http_server"`

	DB     db.SelectorConfig     `json:"db_selector" yaml:"db_selector"`
	Influx influx.SelectorConfig `json:"influx_selector" yaml:"influx_selector"`

	Kafka *kfk.Config `json:"kafka" yaml:"kafka"`
}

func NewConfig() (Config, error) {
	cfg := new(Config)
	err := ReadConfig(cfg, ReadCfgPathFlag())
	if err != nil {
		return Config{}, err
	}
	return *cfg, nil
}

func ReadCfgPathFlag() string {
	p := viper.GetString("cfg")
	if p == "" {
		pflag.StringP("cfg", "c", "config.yml", "Path to config file")
		pflag.Parse()
		err := viper.BindPFlags(pflag.CommandLine)
		if err != nil {
			panic(err)
		}
	}
	return viper.GetString("cfg")
}

func ReadConfig(i interface{}, filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	ext := path.Ext(filePath)
	switch {
	case ext == ".yaml" || ext == ".yml":
		err = yaml.Unmarshal(data, i)
		if err != nil {
			return err
		}
	case ext == ".json":
		err = json.Unmarshal(data, i)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown config format")
	}
	return nil
}
