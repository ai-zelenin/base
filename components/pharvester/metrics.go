package pharvester

import (
	"sync"
	"time"

	"git.pnhub.ru/core/libs/util"
)

type RequestMetrics struct {
	RequestTiming  time.Duration
	ResponseTiming time.Duration
}

type ProxyMetrics struct {
	PingTiming     time.Duration
	RequestMetrics []RequestMetrics
	Error          util.MultiError
	mx             sync.Mutex `json:"-" yaml:"-" gorm:"-"`
}

func (p *ProxyMetrics) AddRequestMetric(rm RequestMetrics) {
	p.mx.Lock()
	defer p.mx.Unlock()
	if p.RequestMetrics == nil {
		p.RequestMetrics = make([]RequestMetrics, 0)
	}
	p.RequestMetrics = append(p.RequestMetrics, rm)
}
