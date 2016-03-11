// Package logging provides a logger, as well as configuration.
package logging

import (
	"github.com/giantswarm/request-context"
)

func NewConfig(level string, color bool) requestcontext.LoggerConfig {
	return requestcontext.LoggerConfig{
		Name:                "inago",
		Level:               level,
		Color:               color,
		IncludeNameInFormat: false,
	}
}

// NewLogger returns a new Logger, given a config.
func NewLogger(config requestcontext.LoggerConfig) *requestcontext.Logger {
	newLogger := requestcontext.MustGetLogger(config)
	return &newLogger
}
