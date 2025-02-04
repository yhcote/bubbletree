// Copyright 2025 Yannick Cote <yhcote@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	charmlog "github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

// DefaultLogger returns a ready to use slog.Logger that has been initialized
// with a charmbracelet logger handler. The debug option sets a preselected
// logger level depending on whether a debug (verbose) mode execution is
// requested.
func DefaultLogger(debug bool) *slog.Logger {
	SetLoggerLevel(debug)
	return log
}

// SetLoggerLevel sets one of two predefined logger levels base on whether
// the program runs in debug mode or not.
func SetLoggerLevel(debug bool) {
	if debug {
		handler.SetLevel(charmlog.DebugLevel)
	} else {
		handler.SetLevel(charmlog.InfoLevel)
	}
}

// GetLoggerOutputName returns the current logger's output file name.
func GetLoggerOutputName() string {
	return output.Name()
}

// CloseLoggerOutput closes the logger's output file when it's a disk file.
func CloseLoggerOutput() error {
	if output != os.Stderr {
		return output.Close()
	}
	return nil
}

var (
	log     *slog.Logger
	handler *charmlog.Logger
	output  *os.File
)

// Initialize the default logger at init time, so that it's ready for caller
// packages.
func init() {
	opts := charmlog.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.StampMicro,
		Formatter:       charmlog.TextFormatter,
		Level:           charmlog.InfoLevel,
	}
	filename := filepath.Join("/", "var", "tmp", fmt.Sprintf("%s-%d.log", filepath.Base(os.Args[0]), os.Getpid()))
	log = loggerWithFile(filename, opts)
}

// loggerWithFile creates a new slog interfaced logger with a charm logger
// handler. The logger writes logs to file: /var/tmp/app-<pid>.log.
func loggerWithFile(filename string, o charmlog.Options) *slog.Logger {
	var err error
	output, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		output = os.Stderr
	}
	handler = charmlog.NewWithOptions(output, o)
	handler.SetColorProfile(termenv.TrueColor)
	return slog.New(handler)
}
