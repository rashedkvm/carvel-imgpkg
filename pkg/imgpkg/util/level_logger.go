// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import goui "github.com/cppforlife/go-cli-ui/ui"

type LogLevel int

type Logger interface {
	Logf(msg string, args ...interface{})
}

type LoggerWithLevels interface {
	Logger

	Errorf(msg string, args ...interface{})
	Warnf(msg string, args ...interface{})
	Debugf(msg string, args ...interface{})
	Tracef(msg string, args ...interface{})
}

type UIWithLevels interface {
	goui.UI

	Errorf(msg string, args ...interface{})
	Warnf(msg string, args ...interface{})
	Debugf(msg string, args ...interface{})
	Tracef(msg string, args ...interface{})
}

const (
	LogTrace LogLevel = iota
	LogDebug LogLevel = iota
	LogWarn  LogLevel = iota
)

func NewUILevelLogger(level LogLevel, ui goui.UI) *UILevelWriter {
	return &UILevelWriter{
		UI:       ui,
		LogLevel: level,
	}
}

func (l ImgpkgLogger) NewLevelLogger(level LogLevel, logger Logger) *LoggerLevelWriter {
	return &LoggerLevelWriter{
		LogLevel: level,
		logger:   logger,
	}
}

type LoggerLevelWriter struct {
	LogLevel LogLevel
	logger   Logger
}

type UILevelWriter struct {
	goui.UI
	LogLevel LogLevel
}

func (l LoggerLevelWriter) Errorf(msg string, args ...interface{}) {
	l.Logf("Error: "+msg, args...)
}

func (l LoggerLevelWriter) Warnf(msg string, args ...interface{}) {
	if l.LogLevel <= LogWarn {
		l.Logf("Warning: "+msg, args...)
	}
}

func (l LoggerLevelWriter) Logf(msg string, args ...interface{}) {
	l.logger.Logf(msg, args...)
}

func (l LoggerLevelWriter) Debugf(msg string, args ...interface{}) {
	if l.LogLevel <= LogDebug {
		l.Logf(msg, args...)
	}
}

func (l LoggerLevelWriter) Tracef(msg string, args ...interface{}) {
	if l.LogLevel == LogTrace {
		l.Logf(msg, args...)
	}
}

func (l UILevelWriter) Errorf(msg string, args ...interface{}) {
	l.Logf("Error: "+msg, args...)
}

func (l UILevelWriter) Warnf(msg string, args ...interface{}) {
	if l.LogLevel <= LogWarn {
		l.Logf("Warning: "+msg, args...)
	}
}

func (l UILevelWriter) Logf(msg string, args ...interface{}) {
	l.BeginLinef(msg, args...)
}

func (l UILevelWriter) Debugf(msg string, args ...interface{}) {
	if l.LogLevel <= LogDebug {
		l.Logf(msg, args...)
	}
}

func (l UILevelWriter) Tracef(msg string, args ...interface{}) {
	if l.LogLevel == LogTrace {
		l.Logf(msg, args...)
	}
}
