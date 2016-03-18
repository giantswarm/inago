// Package logging provides logging.
package logging

import (
	"os"

	"github.com/mattn/go-isatty"
	"golang.org/x/net/context"
)

// Logger provides an interface for other loggers to implement.
type Logger interface {
	// Debug is for logging information pertinent to development and testing.
	Debug(ctx context.Context, f string, v ...interface{})

	// Info is for logging information that is useful for users.
	Info(ctx context.Context, f string, v ...interface{})

	// Notice is for logging information that is useful for users
	// that is of a higher importance than Info.
	Notice(ctx context.Context, f string, v ...interface{})

	// Warning is for logging information that may imply that
	// something untoward has occurred, but that is not a hard error.
	Warning(ctx context.Context, f string, v ...interface{})

	// Error is for logging error messages. Use this in error checking.
	Error(ctx context.Context, f string, v ...interface{})

	// Critical should be used infrequently, and is for cases where
	// human intervention is required to stop catastrophic failure.
	// e.g: losing all units in a cluster, impending nuclear missle launch.
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
