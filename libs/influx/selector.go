package influx

import (
	"context"

	"git.pnhub.ru/core/libs/log"
)

const DefaultInfluxDBKey = "default"

type Selector struct {
	ctx    context.Context
	logger log.Logger
	m      map[string]*Client
}

func NewSelector(ctx context.Context, logger log.Logger, m map[string]*Client) *Selector {
	return &Selector{
		ctx:    ctx,
		logger: logger,
		m:      m,
	}
}

func (s *Selector) Client(keys ...string) *Client {
	var key string
	if len(keys) > 0 {
		key = keys[0]
	} else {
		key = DefaultInfluxDBKey
	}
	x, ok := s.m[key]
	if !ok {
		s.logger.Fatalf("query to undefined influx [%s]", key)
	}
	return x
}
