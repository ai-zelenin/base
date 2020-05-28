// Package kfk contains wrappers and helpers for kafka client
package kfk

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/segmentio/kafka-go"

	"git.pnhub.ru/core/libs/log"
)

type WriterConfigCallback func(config *kafka.WriterConfig)
type ReaderConfigCallback func(config *kafka.ReaderConfig)

type Config struct {
	Brokers []string `json:"brokers" yaml:"brokers"`
}

type Client struct {
	ctx    context.Context
	logger log.Logger
	Cfg    *Config
}

func NewClient(ctx context.Context, logger log.Logger, cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no kafka config")
	}
	k := &Client{ctx: ctx, logger: logger, Cfg: cfg}
	brokenNodes := make(map[string]error)
	for _, broker := range k.Cfg.Brokers {
		_, err := k.APIVersion(broker)
		if err != nil {
			brokenNodes[broker] = err
		} else {
			logger.Infof("kafka node %s is healthy", broker)
		}
	}
	nAll := len(k.Cfg.Brokers)
	nBad := len(brokenNodes)
	hasQuorum := (nAll - nBad) > nAll/2
	if !hasQuorum {
		return nil, fmt.Errorf("has no cuorum: %v", brokenNodes)
	}
	return k, nil
}

func (k *Client) DefaultWriteConfig() *kafka.WriterConfig {
	return &kafka.WriterConfig{
		Brokers:      k.Cfg.Brokers,
		Topic:        "default",
		Balancer:     &kafka.LeastBytes{},
		MaxAttempts:  5,
		BatchSize:    10000,
		BatchTimeout: time.Millisecond * 250,
	}
}

func (k *Client) DefaultReaderConfig() *kafka.ReaderConfig {
	return &kafka.ReaderConfig{
		Brokers:     k.Cfg.Brokers,
		Topic:       "default",
		MaxAttempts: 5,
		MinBytes:    10e3,
		MaxBytes:    200e6,
	}
}

func (k *Client) CreateWriter(cfg *kafka.WriterConfig) *kafka.Writer {
	cfg.Brokers = k.Cfg.Brokers
	if cfg.ErrorLogger == nil {
		cfg.ErrorLogger = k.logger.With(log.LoggerComponentKey, cfg.Topic)
	}
	return kafka.NewWriter(*cfg)
}

func (k *Client) CreateReader(cfg *kafka.ReaderConfig) *kafka.Reader {
	cfg.Brokers = k.Cfg.Brokers
	if cfg.ErrorLogger == nil {
		cfg.ErrorLogger = k.logger.With(log.LoggerComponentKey, cfg.Topic)
	}
	return kafka.NewReader(*cfg)
}

func (k *Client) ControllerConn() (*kafka.Conn, error) {
	var controllerConn *kafka.Conn
	var err error
	for _, broker := range k.Cfg.Brokers {
		controllerConn, err = k.FindController(broker)
		if err != nil {
			_, ok := err.(ErrBrokerNotAvailable)
			if ok {
				k.logger.Infof("node %s not available: %v", broker, err)
				continue
			} else {
				return nil, err
			}
		}
	}
	if controllerConn == nil {
		return nil, fmt.Errorf("cannot find conntroller node")
	}
	return controllerConn, nil
}

func (k *Client) FindController(addr string) (*kafka.Conn, error) {
	conn, err := kafka.DialContext(k.ctx, "tcp", addr)
	if err != nil {
		return nil, ErrBrokerNotAvailable{Cause: err}
	}
	defer conn.Close()
	b, err := conn.Controller()
	if err != nil {
		return nil, err
	}
	return kafka.Dial("tcp", net.JoinHostPort(b.Host, fmt.Sprintf("%d", b.Port)))
}

func (k *Client) APIVersion(addr string) ([]kafka.ApiVersion, error) {
	conn, err := kafka.DialContext(k.ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.ApiVersions()
}
