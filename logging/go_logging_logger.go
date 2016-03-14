package logging

import (
	"fmt"
	"os"

	gologging "github.com/op/go-logging"
	"golang.org/x/net/context"
)

const (
	format = "%{time:2006-01-02 15:04:05} | %{level:-8s} | %{message}"
)

// NewGoLoggingLogger returns a GoLoggingLogger, given a Config.
func NewGoLoggingLogger(config Config) GoLoggingLogger {
	backend := gologging.NewLogBackend(os.Stderr, "", 0)
	backend.Color = config.Color

	leveledBackend := gologging.AddModuleLevel(
		gologging.NewBackendFormatter(
			backend,
			gologging.MustStringFormatter(format),
		),
	)

	level, err := gologging.LogLevel(config.LogLevel)
	if err != nil {
		panic(err)
	}
	leveledBackend.SetLevel(level, config.Name)

	logger := gologging.MustGetLogger(config.Name)
	logger.SetBackend(leveledBackend)

	return GoLoggingLogger{
		logger: logger,
	}
}

// GoLoggingLogger is a Logger based on the go-logging library.
type GoLoggingLogger struct {
	logger *gologging.Logger
}

// Debug logs on the Debug level.
func (gll GoLoggingLogger) Debug(ctx context.Context, f string, v ...interface{}) {
	gll.logger.Debug(gll.formatLog(ctx, f, v...))
}

// Info logs on the Info level.
func (gll GoLoggingLogger) Info(ctx context.Context, f string, v ...interface{}) {
	gll.logger.Info(gll.formatLog(ctx, f, v...))
}

// Notice logs on the Notice level.
func (gll GoLoggingLogger) Notice(ctx context.Context, f string, v ...interface{}) {
	gll.logger.Notice(gll.formatLog(ctx, f, v...))
}

// Warning logs on the Warning level.
func (gll GoLoggingLogger) Warning(ctx context.Context, f string, v ...interface{}) {
	gll.logger.Warning(gll.formatLog(ctx, f, v...))
}

// Error logs on the Error level.
func (gll GoLoggingLogger) Error(ctx context.Context, f string, v ...interface{}) {
	gll.logger.Error(gll.formatLog(ctx, f, v...))
}

// Critical logs on the Critical level.
func (gll GoLoggingLogger) Critical(ctx context.Context, f string, v ...interface{}) {
	gll.logger.Critical(gll.formatLog(ctx, f, v...))
}

// formatLog formats the log record, adding a context if it exists.
func (gll GoLoggingLogger) formatLog(ctx context.Context, f string, v ...interface{}) string {
	if ctx != nil {
		f = fmt.Sprintf("%v: "+f, ctx)
	}

	return fmt.Sprintf(f, v...)
}
