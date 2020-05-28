package log

import (
	"path"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZap(cfg *ZapConfig) (Logger, error) {
	zapCfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:       cfg.Development,
		Encoding:          cfg.Encoding,
		DisableCaller:     cfg.DisableCaller,
		DisableStacktrace: cfg.DisableStacktrace,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.FullCallerEncoder,
		},
		OutputPaths:      cfg.OutputPaths,
		ErrorOutputPaths: cfg.ErrorOutputPaths,
	}
	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}
	return &LoggerWrapper{
		SugaredLogger: logger.Sugar(),
	}, nil
}

// ZapConfig holds variables for zap Logger
type ZapConfig struct {
	// Level is the minimum enabled logging level. Note that this is a dynamic
	// level, so calling InfluxConfig.Level.SetLevel will atomically change the log
	// level of all loggers descended from this config.
	Level string `json:"level" yaml:"level"`
	// Development puts the Logger in development mode, which changes the
	// behavior of DPanicLevel and takes stacktraces more liberally.
	Development bool `json:"development" yaml:"development"`
	// DisableCaller stops annotating logs with the calling function's file
	// name and line number. By default, all logs are annotated.
	DisableCaller bool `json:"disableCaller" yaml:"disableCaller"`
	// DisableStacktrace completely disables automatic stacktrace capturing. By
	// default, stacktraces are captured for WarnLevel and above logs in
	// development and ErrorLevel and above in production.
	DisableStacktrace bool `json:"disableStacktrace" yaml:"disableStacktrace"`
	// Encoding sets the Logger's encoding. Valid values are "json" and
	// "console", as well as any third-party encodings registered via
	// RegisterEncoder.
	Encoding string `json:"encoding" yaml:"encoding"`
	// EncoderConfig sets options for the chosen encoder. See
	// zapcore.EncoderConfig for details.
	EncoderConfig zapcore.EncoderConfig `json:"encoderConfig" yaml:"encoderConfig"`
	// OutputPaths is a list of paths to write logging output to. See Open for
	// details.
	OutputPaths []string `json:"outputPaths" yaml:"outputPaths"`
	// ErrorOutputPaths is a list of paths to write internal Logger errors to.
	// The default is standard error.
	//
	// Note that this setting only affects internal errors; for sample code that
	// sends error-level logs to a different location from info- and debug-level
	// logs, see the package-level AdvancedConfiguration example.
	ErrorOutputPaths []string `json:"errorOutputPaths" yaml:"errorOutputPaths"`
	// InitialFields is a collection of fields to add to the root Logger.
	InitialFields map[string]interface{} `json:"initialFields" yaml:"initialFields"`
}

// DefaultLogger create default Logger struct
func DefaultLogger() Logger {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: true,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   IDECallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return &LoggerWrapper{
		SugaredLogger: logger.Sugar(),
	}
}

func IDECallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	_, f := path.Split(caller.File)
	enc.AppendString(f + ":" + strconv.Itoa(caller.Line))
}

type LoggerWrapper struct {
	*zap.SugaredLogger
}

func (l *LoggerWrapper) With(args ...interface{}) Logger {
	return &LoggerWrapper{
		SugaredLogger: l.SugaredLogger.With(args...),
	}
}
func (l *LoggerWrapper) Print(args ...interface{}) {
	l.SugaredLogger.Info(args...)
}
func (l *LoggerWrapper) Printf(t string, args ...interface{}) {
	l.SugaredLogger.Infof(t, args...)
}
