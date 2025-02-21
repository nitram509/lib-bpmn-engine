package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"strings"

	"github.com/hashicorp/go-hclog"
)

type HcLogger struct {
	slog *slog.Logger
	name string
	args []interface{}
	skip int
}

func NewHcLog(logger *slog.Logger, skip int) *HcLogger {
	return &HcLogger{
		slog: logger,
		skip: skip,
	}
}

func (l *HcLogger) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.Trace:
		l.Trace(msg, args)
	case hclog.Debug:
		l.Debug(msg, args)
	case hclog.Info:
		l.Info(msg, args)
	case hclog.Warn:
		l.Warn(msg, args)
	case hclog.Error:
		l.Error(msg, args)
	case hclog.Off:
		return
	}
}

func (l *HcLogger) argsToAttrs(args ...interface{}) []slog.Attr {
	attrSize := len(args) + len(l.args)
	if l.name != "" {
		attrSize += 1
	}
	attrs := make([]slog.Attr, attrSize)
	for i := range args {
		if t, ok := args[i].([]interface{}); ok {
			if len(t) == 0 {
				continue
			}
			attrs = append(attrs, slog.Attr{
				Key:   fmt.Sprintf("%s", t[0]),
				Value: slog.AnyValue(t[1]),
			})
		}
	}
	for i := range l.args {
		if t, ok := l.args[i].([]interface{}); ok {
			if len(t) == 0 {
				continue
			}
			attrs = append(attrs, slog.Attr{
				Key:   fmt.Sprintf("%s", t[0]),
				Value: slog.AnyValue(t[1]),
			})
		}
	}
	if l.name != "" {
		attrs = append(attrs, slog.String("logger", l.name))
	}
	return attrs
}

func (l *HcLogger) Trace(msg string, args ...interface{}) {
	logSkipCallers(context.Background(), l.skip, LevelTrace, msg, l.argsToAttrs(args))
}

func (l *HcLogger) Debug(msg string, args ...interface{}) {
	logSkipCallers(context.Background(), l.skip, slog.LevelDebug, msg, l.argsToAttrs(args))
}

func (l *HcLogger) Info(msg string, args ...interface{}) {
	logSkipCallers(context.Background(), l.skip, slog.LevelInfo, msg, l.argsToAttrs(args))
}

func (l *HcLogger) Warn(msg string, args ...interface{}) {
	logSkipCallers(context.Background(), l.skip, slog.LevelWarn, msg, l.argsToAttrs(args))
}

func (l *HcLogger) Error(msg string, args ...interface{}) {
	logSkipCallers(context.Background(), l.skip, slog.LevelError, msg, l.argsToAttrs(args))
}

func (l *HcLogger) IsTrace() bool {
	return l.slog.Enabled(context.Background(), LevelTrace)
}

func (l *HcLogger) IsDebug() bool {
	return l.slog.Enabled(context.Background(), slog.LevelDebug)
}

func (l *HcLogger) IsInfo() bool {
	return l.slog.Enabled(context.Background(), slog.LevelInfo)
}

func (l *HcLogger) IsWarn() bool {
	return l.slog.Enabled(context.Background(), slog.LevelWarn)
}

func (l *HcLogger) IsError() bool {
	return l.slog.Enabled(context.Background(), slog.LevelError)
}

func (l *HcLogger) ImpliedArgs() []interface{} {
	return l.args
}

func (l *HcLogger) With(args ...interface{}) hclog.Logger {
	logger := NewHcLog(l.slog.With(args...), l.skip)
	logger.name = l.name
	logger.args = append(l.args, args)
	return logger
}

func (l *HcLogger) Name() string {
	return l.name
}

func (l *HcLogger) Named(name string) hclog.Logger {
	logger := NewHcLog(l.slog, l.skip)
	logger.name = name
	logger.args = l.args
	return logger
}

func (l *HcLogger) ResetNamed(name string) hclog.Logger {
	logger := NewHcLog(l.slog, l.skip)
	logger.name = fmt.Sprintf("%s-%s", l.name, name)
	logger.args = l.args
	return logger
}

func (l *HcLogger) SetLevel(level hclog.Level) {
	// no-op
}

func (l *HcLogger) GetLevel() hclog.Level {
	if l.slog.Enabled(context.Background(), slog.LevelInfo) {
		return hclog.Info
	}
	if l.slog.Enabled(context.Background(), slog.LevelError) {
		return hclog.Error
	}
	if l.slog.Enabled(context.Background(), slog.LevelWarn) {
		return hclog.Warn
	}
	if l.slog.Enabled(context.Background(), slog.LevelDebug) {
		return hclog.Debug
	}
	if l.slog.Enabled(context.Background(), LevelTrace) {
		return hclog.Trace
	}
	return hclog.Off
}

func (l *HcLogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(l.StandardWriter(opts), "", log.LstdFlags)
}

func (l *HcLogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return &stdlogAdapter{
		log:        l,
		forceLevel: opts.ForceLevel,
	}
}

// Provides a io.Writer to shim the data out of *log.Logger
// and back into our Logger. This is basically the only way to
// build upon *log.Logger.
type stdlogAdapter struct {
	log        hclog.Logger
	forceLevel hclog.Level
}

// Take the data, infer the levels if configured, and send it through
// a regular Logger.
func (s *stdlogAdapter) Write(data []byte) (int, error) {
	str := string(bytes.TrimRight(data, " \t\n"))

	if s.forceLevel != hclog.NoLevel {
		// Use pickLevel to strip log levels included in the line since we are
		// forcing the level
		_, str := s.pickLevel(str)

		// Log at the forced level
		s.dispatch(str, s.forceLevel)
	} else {
		s.log.Info(str)
	}

	return len(data), nil
}

func (s *stdlogAdapter) dispatch(str string, level hclog.Level) {
	switch level {
	case hclog.Trace:
		s.log.Trace(str)
	case hclog.Debug:
		s.log.Debug(str)
	case hclog.Info:
		s.log.Info(str)
	case hclog.Warn:
		s.log.Warn(str)
	case hclog.Error:
		s.log.Error(str)
	default:
		s.log.Info(str)
	}
}

// Detect, based on conventions, what log level this is.
func (s *stdlogAdapter) pickLevel(str string) (hclog.Level, string) {
	switch {
	case strings.HasPrefix(str, "[DEBUG]"):
		return hclog.Debug, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[TRACE]"):
		return hclog.Trace, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[INFO]"):
		return hclog.Info, strings.TrimSpace(str[6:])
	case strings.HasPrefix(str, "[WARN]"):
		return hclog.Warn, strings.TrimSpace(str[6:])
	case strings.HasPrefix(str, "[ERROR]"):
		return hclog.Error, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[ERR]"):
		return hclog.Error, strings.TrimSpace(str[5:])
	default:
		return hclog.Info, str
	}
}
