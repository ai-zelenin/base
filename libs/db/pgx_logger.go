package db

import (
	"context"

	"github.com/jackc/pgx/v4"

	"git.pnhub.ru/core/libs/log"
)

type PGXLogger struct {
	logger log.Logger
}

func (l *PGXLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	switch level {
	case pgx.LogLevelTrace:
		l.logger.Debug(msg, data)
	case pgx.LogLevelDebug:
		l.logger.Debug(msg, data)
	case pgx.LogLevelInfo:
		l.logger.Info(msg, data)
	case pgx.LogLevelWarn:
		l.logger.Warn(msg, data)
	case pgx.LogLevelError:
		l.logger.Error(msg, data)
	}
}
