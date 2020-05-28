// Package base contain mostly common application logic and external deps
package base

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"go.uber.org/fx"

	"git.pnhub.ru/core/libs/log"
)

var ErrNilDependency = func(i interface{}) error {
	return fmt.Errorf("dependency not found %s", reflect.TypeOf(i).String())
}

type ProviderFunction interface{}

type Application struct {
	Ctx           context.Context
	Cancel        context.CancelFunc
	FxApplication *fx.App
	logger        log.Logger
	providers     []interface{}
}

func NewApplication(providers []ProviderFunction) *Application {
	var pp = make([]interface{}, 0)
	for _, p := range providers {
		pp = append(pp, p)
	}
	ctx, c := context.WithCancel(context.Background())
	return &Application{
		providers: pp,
		Ctx:       ctx,
		Cancel:    c,
		logger:    log.DefaultLogger(),
	}
}

func (a *Application) Start(invoker interface{}) {
	a.providers = append(a.providers, func() context.Context { return a.Ctx })
	a.FxApplication = fx.New(
		fx.Provide(a.providers...),
		fx.Invoke(invoker),
	)

	go a.ListenSignals()

	startCtx, cancel := context.WithTimeout(a.Ctx, fx.DefaultTimeout)
	defer cancel()
	err := a.FxApplication.Start(startCtx)
	if err != nil {
		a.logger.Fatal(err)
	}
}

func (a *Application) Stop() {
	stopCtx, cancel := context.WithTimeout(a.Ctx, fx.DefaultTimeout)
	defer cancel()
	err := a.FxApplication.Stop(stopCtx)
	if err != nil {
		a.logger.Fatal(err)
	}
	a.FxApplication = nil
}

func (a *Application) ListenSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for sig := range signals {
		a.logger.Infof("income signal %s", sig)
		a.Stop()
		a.Cancel()
		return
	}
}
