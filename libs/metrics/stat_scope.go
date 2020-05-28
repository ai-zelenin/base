package metrics

import (
	"context"
	"time"

	"github.com/uber-go/tally"

	"git.pnhub.ru/core/libs/log"
)

type Reporter interface {
	tally.StatsReporter
}

func NewStatScope(ctx context.Context, logger log.Logger, reporter Reporter) tally.Scope {
	options := tally.ScopeOptions{
		Reporter:  reporter,
		Separator: "-",
	}

	scope, closer := tally.NewRootScope(options, 1*time.Second)

	go func() {
		for range ctx.Done() {
			err := closer.Close()
			if err != nil {
				logger.Error(err)
			}
		}
	}()
	return scope
}
