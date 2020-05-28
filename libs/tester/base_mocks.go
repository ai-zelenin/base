// Package tester contains helpers for writing unit tests
package tester

import (
	"context"

	"git.pnhub.ru/core/libs/log"
)

func MockContext() context.Context {
	return context.Background()
}

func MockLogger() log.Logger {
	return log.DefaultLogger().With("instance", "test")
}
