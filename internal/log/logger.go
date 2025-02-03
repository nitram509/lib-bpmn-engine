package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

func Init() {
	logLevel := os.Getenv("LOG_LEVEL")
	var logger *slog.Logger
	logOptions := &slog.HandlerOptions{
		AddSource: true,
	}
	if logLevel != "" {
		switch logLevel {
		case "DEBUG":
			logOptions.Level = slog.LevelDebug
		case "WARN":
			logOptions.Level = slog.LevelWarn
		case "ERROR":
			logOptions.Level = slog.LevelError
		default:
			logOptions.Level = slog.LevelInfo
		}
	}
	// add option to use json handler
	logger = slog.New(
		slog.NewTextHandler(os.Stdout, logOptions),
	)
	slog.SetDefault(logger)
}

func Error(msg string, a ...any) {
	logSkipCallers(context.Background(), 1, slog.LevelError, fmt.Sprintf(msg, a...), nil)
}

func Errorf(ctx context.Context, msg string, a ...any) {
	logSkipCallers(ctx, 1, slog.LevelError, fmt.Sprintf(msg, a...), nil)
}

func Errora(ctx context.Context, msg string, attrs ...slog.Attr) {
	logSkipCallers(ctx, 1, slog.LevelError, msg, attrs)
}

func Info(msg string, a ...any) {
	logSkipCallers(context.Background(), 1, slog.LevelInfo, fmt.Sprintf(msg, a...), nil)
}
func Infof(ctx context.Context, msg string, a ...any) {
	logSkipCallers(ctx, 1, slog.LevelInfo, fmt.Sprintf(msg, a...), nil)
}

func Infoa(ctx context.Context, msg string, attrs ...slog.Attr) {
	logSkipCallers(ctx, 1, slog.LevelInfo, msg, attrs)
}

func Debug(msg string, a ...any) {
	logSkipCallers(context.Background(), 1, slog.LevelDebug, fmt.Sprintf(msg, a...), nil)
}

func Debugf(ctx context.Context, msg string, a ...any) {
	logSkipCallers(ctx, 1, slog.LevelDebug, fmt.Sprintf(msg, a...), nil)
}

func Debuga(ctx context.Context, msg string, attrs ...slog.Attr) {
	logSkipCallers(ctx, 1, slog.LevelDebug, msg, attrs)
}

func Warn(msg string, a ...any) {
	logSkipCallers(context.Background(), 1, slog.LevelWarn, fmt.Sprintf(msg, a...), nil)
}

func Warnf(ctx context.Context, msg string, a ...any) {
	logSkipCallers(ctx, 1, slog.LevelWarn, fmt.Sprintf(msg, a...), nil)
}

func Warna(ctx context.Context, msg string, attrs ...slog.Attr) {
	logSkipCallers(ctx, 1, slog.LevelWarn, msg, attrs)
}

func getContextArgs(ctx context.Context) []slog.Attr {
	// baseAttrs := []slog.Attr{
	// 	slog.Group("user",
	// 		slog.String("id", appcontext.GetUserId(ctx)),
	// 	),
	// }
	// return baseAttrs
	return []slog.Attr{}
}

func logSkipCallers(ctx context.Context, skip int, level slog.Level, msg string, attrs []slog.Attr) {
	var pcs [1]uintptr
	runtime.Callers(skip+2, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.AddAttrs(append(getContextArgs(ctx), attrs...)...)
	_ = slog.Default().Handler().Handle(context.Background(), r)
}
