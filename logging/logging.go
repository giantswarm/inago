// Package logging provides a logger, as well as configuration.
package logging

import (
	"github.com/giantswarm/request-context"
)

var (
	logger *requestcontext.Logger

	// The default log level to use.
	logLevel = "INFO"
	// Whether to use color, by default.
	useColor = false
)

// SetLogLevel sets the log level of the logger to be created.
func SetLogLevel(level string) {
	logLevel = level
}

// UseColor sets whether to use color in log output.
func UseColor(color bool) {
	useColor = color
}

// GetLogger creates the default logger if it doesn't exist,
// returning the logger.
func GetLogger() *requestcontext.Logger {
	if logger == nil {
		l := requestcontext.MustGetLogger(
			requestcontext.LoggerConfig{
				Name:                "inago",
				Level:               logLevel,
				Color:               useColor,
				IncludeNameInFormat: false,
			},
		)
		logger = &l
	}

	return logger
}
