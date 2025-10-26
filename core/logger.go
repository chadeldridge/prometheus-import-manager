package core

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	quiet  bool
	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr
)

type Logger struct {
	DebugMode bool
	*log.Logger
}

func NewLogger(out io.Writer, prefix string, flag int, debug bool) *Logger {
	return &Logger{DebugMode: debug, Logger: log.New(out, prefix, flag)}
}

func (l *Logger) Quiet() {
	quiet = true
}

/*
func getLogID() string {
	logID := strings.Replace(uuid.New().String(), "-", "", -1)
	return logID
}
*/

// Debug checks that DebugMode is enabled before using Logger.Print. "[DEBUG] " will be prepended
// to the line.
func (l *Logger) Debug(a ...any) {
	if quiet {
		return
	}

	if l.DebugMode {
		l.Print(append([]any{"[DEBUG] "}, a...)...)
	}
}

// Debug checks that DebugMode is enabled before using Logger.Printf. "[DEBUG] " will be prepended
// to the line.
func (l *Logger) Debugf(format string, a ...any) {
	if quiet {
		return
	}

	if l.DebugMode {
		l.Printf("[DEBUG] "+format, a...)
	}
}

// Print to log functions.

// Print is a wrapper for log.Print.
func (l *Logger) Print(a ...any) {
	if quiet {
		return
	}

	l.Logger.Print(a...)
}

// Printf is a wrapper for log.Printf.
func (l *Logger) Printf(format string, a ...any) {
	if quiet {
		return
	}

	l.Logger.Printf(format, a...)
}

// Print to console functions

// PrintOut uses Fprintln to Core.Stdout io.Writer.
func (l *Logger) PrintOut(a ...any) {
	if quiet {
		return
	}

	fmt.Fprintln(Stdout, a...)
}

// PrintOutf uses Fprintf to Core.Stdout io.Writer.
func (l *Logger) PrintOutf(format string, a ...any) {
	if quiet {
		return
	}

	fmt.Fprintf(Stdout, format, a...)
}

// PrintErr uses Fprintln to Core.Stderr io.Writer.
func (l *Logger) PrintErr(a ...any) {
	fmt.Fprintln(Stderr, a...)
}

// PrintErrf uses Fprintf to Core.Stderr io.Writer.
func (l *Logger) PrintErrf(format string, a ...any) {
	fmt.Fprintf(Stderr, format, a...)
}
