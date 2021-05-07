package log

// Reasoning for log pkg
// 1. Logger should not be set as a handler attribute as it is not request-scoped

import (
	"context"
	"strings"

	// Using zap instead of go-kit/log because go-kit/log/level requires a go-kit/log logger,
	// and thought a logger struct was overkill because I didn't not want to recreate the zap methods
	// (zap provides no interfaces).
	"go.uber.org/zap"
)

type ctxLogKey struct{}

var logger *zap.Logger
var cfg zap.Config

func init() {
	cfg = zap.NewProductionConfig()
	l, err := cfg.Build(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	if err != nil {
		panic(err.Error())
	}

	var fields []zap.Field

	if len(fields) > 0 {
		l = l.With(fields...)
	}

	logger = l.Named("logger")
}

// Log writes log to output.
func Log() *zap.SugaredLogger {
	return logger.Sugar()
}

// SetLevel sets the level of the logger.
func SetLevel(lvl string) {
	switch strings.ToUpper(lvl) {
	case "DEBUG":
		cfg.Level.SetLevel(zap.DebugLevel)
	}
}

// Sync triggers a sync of the underlying zap Logger.
func Sync(ctx context.Context) {
	FromContext(ctx).Sync()
	return
}

// WithCtx is a convenience method for logging with contexts using a SugaredLogger
func WithCtx(ctx context.Context) *zap.SugaredLogger {
	return FromContext(ctx).Sugar()
}

// FromContext returns a zap.Logger with additional optional fields.
func FromContext(ctx context.Context, fields ...zap.Field) *zap.Logger {
	if ctx == nil {
		ctx = context.Background()
	}

	l, ok := ctx.Value(ctxLogKey{}).(*zap.Logger)
	if !ok || l == nil {
		l = logger
	}

	if len(fields) > 0 {
		return l.With(fields...)
	}

	return l
}

// ToContext pushes a new logger, with additional optional fields, into the context.
func ToContext(ctx context.Context, base *zap.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, ctxLogKey{}, base)
}

// AddFields is sugar for updating the in-context logger with additional fields.
func AddFields(ctx context.Context, fields ...zap.Field) context.Context {
	l := FromContext(ctx, fields...)
	return ToContext(ctx, l)
}
