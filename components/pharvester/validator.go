package pharvester

import (
	"context"
	"fmt"
	"sync"
	"time"

	"git.pnhub.ru/core/libs/base"
	"git.pnhub.ru/core/libs/log"
)

var ValidationSchemas = []string{
	"http",
	"https",
}

const (
	ProxyStatusOk       = 1
	ProxyStatusFailed   = 0
	ProxyStatusNeedScan = -1
)

var ErrorScores = map[int]int64{
	ErrCodeUntyped:                 -1,
	ErrCodePacketLoss:              -10,
	ErrCodeRequestFail:             -20,
	ErrCodeBadResponse:             -20,
	ErrCodeCannotCreateProxyDialer: -20,
	ErrCodeInvalidProxyURL:         0,
}

var MetricsScores = []struct {
	Min   time.Duration
	Max   time.Duration
	Score int64
}{
	{
		Min:   0,
		Max:   time.Second * 2,
		Score: 30,
	},
	{
		Min:   time.Second * 2,
		Max:   time.Second * 3,
		Score: 20,
	},
	{
		Min:   time.Second * 3,
		Max:   time.Second * 5,
		Score: 5,
	},
	{
		Min:   time.Second * 5,
		Max:   time.Second * 100,
		Score: 1,
	},
}

type Validator struct {
	ctx     context.Context
	logger  log.Logger
	cfg     *ValidatorConfig
	storage *Storage
	input   chan *Proxy
}

func NewValidator(ctx context.Context, logger log.Logger, cfg *Config, storage *Storage) (*Validator, error) {
	if cfg.ValidatorConfig == nil {
		return nil, base.ErrNilDependency(cfg.ValidatorConfig)
	}
	return &Validator{
		ctx:     ctx,
		logger:  logger,
		cfg:     cfg.ValidatorConfig,
		storage: storage,
		input:   make(chan *Proxy, cfg.ValidatorConfig.Threads),
	}, nil
}

func (v *Validator) Start() error {
	go v.serveInput()
	for {
		pl, err := v.storage.LoadValidationBunch(1000)
		if err != nil {
			return err
		}
		v.logger.Debugf("validation bunch %d", len(pl))
		if len(pl) == 0 {
			time.Sleep(time.Second * 5)
		}
		for _, proxy := range pl {
			v.input <- proxy
		}
	}
}
func (v *Validator) serveInput() {
	var semaphore = make(chan struct{}, v.cfg.Threads)
	for {
		select {
		case <-v.ctx.Done():
			return
		case pr, ok := <-v.input:
			if !ok {
				return
			}
			semaphore <- struct{}{}
			go func(proxy *Proxy) {
				defer func() {
					<-semaphore
				}()
				v.Validate(proxy)
			}(pr)
		}
	}
}

func (v *Validator) Validate(proxy *Proxy) {
	v.logger.Debugf("Validating %s", proxy.URL())
	proxy.Metrics = new(ProxyMetrics)

	defer func() {
		v.CalcErrors(proxy)
		v.CalcMetrics(proxy)
		v.CalcStatus(proxy)
		proxy.LastCheck = time.Now().UTC()
		proxy.CheckNumber++
		err := v.storage.Save(proxy)
		if err != nil {
			v.logger.Error(err)
		}
	}()

	err := proxy.InitClient(v.cfg.Timeout)
	if err != nil {
		proxy.Metrics.Error.Add(err, ErrCodeInvalidProxyURL)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err = proxy.CheckPing()
		if err != nil {
			er, ok := err.(*Error)
			if ok {
				proxy.Metrics.Error.Add(er, er.Code)
			} else {
				proxy.Metrics.Error.Add(er, ErrCodeUntyped)
			}
		}
	}()

	for _, schema := range ValidationSchemas {
		for i := 0; i < v.cfg.NumberOfRequests; i++ {
			wg.Add(1)
			go func(s string) {
				defer wg.Done()
				err = proxy.CheckProxy(fmt.Sprintf("%s://%s%s", s, v.cfg.Host, v.cfg.Path))
				if err != nil {
					er, ok := err.(*Error)
					if ok {
						proxy.Metrics.Error.Add(er, er.Code)
					} else {
						proxy.Metrics.Error.Add(er, ErrCodeUntyped)
					}
				}
			}(schema)
		}
	}

	wg.Wait()

}

func (v *Validator) CalcMetrics(proxy *Proxy) {
	var allRequest time.Duration
	var allResponse time.Duration
	var n float64

	for _, m := range proxy.Metrics.RequestMetrics {
		allRequest += m.RequestTiming
		allResponse += m.ResponseTiming
		n++
	}

	proxy.AvgRequestTiming = time.Duration(float64(allRequest) / n)
	if proxy.AvgRequestTiming < 0 {
		proxy.AvgRequestTiming = 0
	}
	proxy.AvgResponseTiming = time.Duration(float64(allResponse) / n)
	if proxy.AvgResponseTiming < 0 {
		proxy.AvgResponseTiming = 0
	}
	proxy.AvgPingTiming = proxy.Metrics.PingTiming

	for _, ms := range MetricsScores {
		t := proxy.AvgResponseTiming
		if t > ms.Min && t < ms.Max {
			proxy.Score += ms.Score
			break
		}
	}
}

func (v *Validator) CalcErrors(proxy *Proxy) {
	for code, score := range ErrorScores {
		err := proxy.Metrics.Error.CheckByKey(code)
		if err != nil {
			v.logger.Debugf("Code:%d %s %v", code, proxy.URL(), err)
			proxy.Score += score
		}
	}
}
func (v *Validator) CalcStatus(proxy *Proxy) {
	err := proxy.Metrics.Error.CheckByKey(ErrCodeInvalidProxyURL)
	if err != nil {
		proxy.Status = ProxyStatusNeedScan
		return
	}
	switch {
	case proxy.Score > 0:
		proxy.Status = ProxyStatusOk
	case proxy.Score <= 0:
		proxy.Status = ProxyStatusFailed
	}
}
