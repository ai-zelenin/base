package pharvester

import (
	"context"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"git.pnhub.ru/core/libs/log"
)

type Harvester struct {
	ctx       context.Context
	logger    log.Logger
	cfg       *Config
	crawler   *Crawler
	parser    *Parser
	validator *Validator
	storage   *Storage
}

func NewHarvester(ctx context.Context, logger log.Logger, cfg *Config, crawler *Crawler, parser *Parser, validator *Validator, storage *Storage) *Harvester {
	return &Harvester{
		ctx:       ctx,
		logger:    log.ForkLogger(logger),
		cfg:       cfg,
		crawler:   crawler,
		parser:    parser,
		validator: validator,
		storage:   storage,
	}
}
func (h *Harvester) ImportFromFile() error {
	data, err := ioutil.ReadFile(h.cfg.ImportFilePath)
	if err != nil {
		return err
	}
	pl := make([]*Proxy, 0)

	err = yaml.Unmarshal(data, &pl)
	if err != nil {
		return err
	}
	for _, p := range pl {
		err = h.storage.Save(p)
		if err != nil {
			h.logger.Warn(err)
		}
	}
	return nil
}

func (h *Harvester) Harvest() error {
	if h.cfg.ImportFilePath != "" {
		err := h.ImportFromFile()
		if err != nil {
			return err
		}
	}

	for _, sourceConfig := range h.cfg.SourceConfigs {
		if sourceConfig.Skip {
			continue
		}
		err := h.crawler.CrawlSourceURL(sourceConfig)
		if err != nil {
			return err
		}
	}

	return h.validator.Start()
}
