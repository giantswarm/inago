package logging

import (
	"github.com/giantswarm/request-context"
)

// NewRequestContextLogger returns a RequestContextLogger, given a Config.
func NewRequestContextLogger(config Config) RequestContextLogger {
	requestContextConfig := requestcontext.LoggerConfig{
		Name:                config.Name,
		Level:               config.LogLevel,
		Color:               config.Color,
		IncludeNameInFormat: false,
	}

	rcl := RequestContextLogger{
		logger: requestcontext.MustGetLogger(requestContextConfig),
	}

	return rcl
}

// RequestContextLogger is a Logger based on the request-context library.
type RequestContextLogger struct {
	logger requestcontext.Logger
}

// contextToCtx takes a logging.Context, and returns a requestcontext.Ctx.
func contextToCtx(context Context) requestcontext.Ctx {
	if context == nil {
		return nil
	}

	ctx := requestcontext.Ctx{}
	for key, value := range context {
		ctx[key] = value
	}

	return ctx
}

// Debug logs on the Debug level.
func (rcl RequestContextLogger) Debug(context Context, f string, v ...interface{}) {
	rcl.logger.Debug(contextToCtx(context), f, v...)
}

// Info logs on the Info level.
func (rcl RequestContextLogger) Info(context Context, f string, v ...interface{}) {
	rcl.logger.Info(contextToCtx(context), f, v...)
}

// Notice logs on the Notice level.
func (rcl RequestContextLogger) Notice(context Context, f string, v ...interface{}) {
	rcl.logger.Notice(contextToCtx(context), f, v...)
}

// Warning logs on the Warning level.
func (rcl RequestContextLogger) Warning(context Context, f string, v ...interface{}) {
	rcl.logger.Warning(contextToCtx(context), f, v...)
}

// Error logs on the Error level.
func (rcl RequestContextLogger) Error(context Context, f string, v ...interface{}) {
	rcl.logger.Error(contextToCtx(context), f, v...)
}

// Critical logs on the Critical level.
func (rcl RequestContextLogger) Critical(context Context, f string, v ...interface{}) {
	rcl.logger.Critical(contextToCtx(context), f, v...)
}
