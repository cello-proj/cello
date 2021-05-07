package log

// Reasoning for log pkg
// 1. Logger should not be set as a handler attribute as it is not request-scoped

import (
	"context"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type ctxLogKey struct{}

// SetLevel sets the level of the logger.
func SetLevel(logger *log.Logger, lvl string) {
	switch strings.ToUpper(lvl) {
	case "DEBUG":
		*logger = level.NewFilter(*logger, level.AllowDebug())
	default:
		*logger = level.NewFilter(*logger, level.AllowInfo())
	}
}

// WithCtx is a convenience method for logging using logger in context.
func WithCtx(ctx context.Context) log.Logger {
	return FromContext(ctx)
}

// FromContext returns a log.Logger with additional optional fields.
func FromContext(ctx context.Context, fields ...interface{}) log.Logger {
	if ctx == nil {
		ctx = context.Background()
	}

	l, ok := ctx.Value(ctxLogKey{}).(log.Logger)
	if !ok || l == nil {
		l = GetLogger(ctx)
	}

	if len(fields) > 0 {
		l = log.With(l, fields...)
		return l
	}
	return l
}

// ToContext pushes a new logger, with additional optional fields, into the context.
func ToContext(ctx context.Context, base log.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxLogKey{}, base)
}

// AddFields updates the in-context logger with additional fields.
func AddFields(ctx context.Context, fields ...interface{}) context.Context {
	l := FromContext(ctx, fields...)
	return ToContext(ctx, l)
}

// GetLogger returns the logger from context or creates a new logger if not set in context.
func GetLogger(ctx context.Context) log.Logger {
	logger, ok := ctx.Value(ctxLogKey{}).(log.Logger)
	if !ok || logger == nil {
		logger = log.With(log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)), "ts", log.DefaultTimestampUTC)
	}
	return logger
}
