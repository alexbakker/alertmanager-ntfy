package main

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/copystructure"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	// Register a special copy handler for zap.AtomicLevel where the same
	// instance is returned instead of creating a copy. This hack is needed
	// because zap.AtomicLevel contains a private *atomic.Int32.
	copystructure.Copiers[reflect.TypeOf(zap.AtomicLevel{})] = func(i interface{}) (interface{}, error) {
		return i, nil
	}
}

func newLogger(config *zap.Config) (*zap.Logger, error) {
	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("build zap logger: %w", err)
	}

	return logger, nil
}

func getDefaultLogConfig(level zapcore.Level) *zap.Config {
	return &zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: level == zapcore.DebugLevel,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
}
