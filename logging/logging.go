// Package logging provides logging.
package logging

import (
	"os"

	"github.com/mattn/go-isatty"
	"golang.org/x/net/context"
)

// Logger provides an interface for other loggers to implement.
type Logger interface {
	Debug(ctx context.Context, f string, v ...interface{})
	Info(ctx context.Context, f string, v ...interface{})
	Notice(ctx context.Context, f string, v ...interface{})
	Warning(ctx context.Context, f string, v ...interface{})
	Error(ctx context.Context, f string, v ...interface{})
	Critical(ctx context.Context, f string, v ...interface{})
}

// Config provides a configuration for loggers.
type Config struct {
	// Name determines the name of the logger.
	Name string
	// Level determines the level of the logger.
	LogLevel string
	// Color determines whether logs are printed with color, where supported.
	Color bool
}

// DefaultConfig returns a Config set by best effort.
func DefaultConfig() Config {
	return Config{
		Name:     "inago",
		LogLevel: "INFO",
		Color:    isatty.IsTerminal(os.Stderr.Fd()),
	}
}

// NewLogger returns a Logger.
func NewLogger(config Config) Logger {
	return NewGoLoggingLogger(config)
}
