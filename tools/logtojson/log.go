package main

import (
	"os"

	"io"

	"fmt"
	"path/filepath"

	"github.com/btcsuite/btclog"
	"github.com/jrick/logrotate/rotator"
)

// logWriter implements an io.Writer that outputs to both standard output and
// the write-end pipe of an initialized log rotator.
type logWriter struct{}

func (logWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write(p)
	logRotatorPipe.Write(p)
	return len(p), nil
}

// Loggers per subsystem.  A single backend hublogs is created and all subsytem
// hublogss created from it will write to the backend.  When adding new
// subsystems, add the subsystem hublogs variable here and to the
// subsystemLoggers map.
//
// Loggers can not be used before the log rotator has been initialized with a
// log file.  This must be performed early during application startup by
// calling initLogRotator.
var (
	// backendLog is the logging backend used to create all subsystem
	// hublogss.  The backend must not be used before the log rotator has
	// been initialized, or data races and/or nil pointer dereferences will
	// occur.
	backendLog = btclog.NewBackend(logWriter{})

	// logRotator is one of the logging outputs.  It should be closed on
	// application shutdown.
	logRotator *rotator.Rotator

	// logRotatorPipe is the write-end pipe for writing to the log rotator.
	// It is written to by the Write method of the logWriter type.
	logRotatorPipe *io.PipeWriter

	mainLog = backendLog.Logger("MAIN")
)

// Initialize package-global hublogs variables.
func init() {
}

// subsystemLoggers maps each subsystem identifier to its associated hublogs.
var subsystemLoggers = map[string]btclog.Logger{
	"BACKEND": mainLog,
}

// initLogRotator initializes the logging rotator to write logs to logFile and
// create roll files in the same directory.  It must be called before the
// package-global log rotator variables are used.
func initLogRotator(logFile string) func() {
	logDir, _ := filepath.Split(logFile)
	err := os.MkdirAll(logDir, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log directory: %v\n", err)
		os.Exit(1)
	}
	r, err := rotator.New(logFile, 10*1024, false, 3)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file rotator: %v\n", err)
		os.Exit(1)
	}

	pr, pw := io.Pipe()
	go r.Run(pr)

	logRotator = r
	logRotatorPipe = pw

	return func() {
		if logRotator != nil {
			logRotator.Close()
		}
	}
}

// setLogLevel sets the logging level for provided subsystem.  Invalid
// subsystems are ignored.  Uninitialized subsystems are dynamically created as
// needed.
func setLogLevel(subsystemID string, logLevel string) {
	// Ignore invalid subsystems.
	hublogs, ok := subsystemLoggers[subsystemID]
	if !ok {
		return
	}

	// Defaults to info if the log level is invalid.
	level, _ := btclog.LevelFromString(logLevel)
	hublogs.SetLevel(level)
}

// setLogLevels sets the log level for all subsystem hublogss to the passed
// level. It also dynamically creates the subsystem hublogss as needed, so it
// can be used to initialize the logging system.
func setLogLevels(logLevel string) {
	// Configure all sub-systems with the new logging level.  Dynamically
	// create hublogss as needed.
	for subsystemID := range subsystemLoggers {
		setLogLevel(subsystemID, logLevel)
	}
}

// logClosure is used to provide a closure over expensive logging operations so
// don't have to be performed when the logging level doesn't warrant it.
type logClosure func() string

// String invokes the underlying function and returns the result.
func (c logClosure) String() string {
	return c()
}

// newLogClosure returns a new closure over a function that returns a string
// which itself provides a Stringer interface so that it can be used with the
// logging system.
func newLogClosure(c func() string) logClosure {
	return logClosure(c)
}
