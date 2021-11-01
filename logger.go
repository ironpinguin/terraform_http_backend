package main

import (
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

// GetLogger returns the logger instance.
// This instance is the entry point for all logging
func GetLogger() *logrus.Logger {
	return logger
}

// SetLogger sets the logger instance
// This is useful in testing as the logger can be overridden
// with a test logger
func SetLogger(l *logrus.Logger) {
	logger = l
}
