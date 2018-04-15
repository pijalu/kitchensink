package quietlog

//go:generate mockgen -source=quietlog.go -package mocks -destination=../mocks/quietlog.go

import (
	"log"
	"os"
)

// Quieter interface allows to retrieve Quiet flag value
type Quieter interface {
	// Quiet returns true if the logger should be quiet
	Quiet() bool
}

// Printfer defines printf method
type Printfer interface {
	Printf(format string, v ...interface{})
}

// Fatalfer defines printf method
type Fatalfer interface {
	Fatalf(format string, v ...interface{})
}

// Logger define the set of method a logger should implement
type Logger interface {
	Printfer
	Fatalfer
}

// QuietLogger define an object with quiet logging option
type QuietLogger struct {
	q Quieter
	l Logger
}

// DefaultLogger returns a default log.Logger based logger
func DefaultLogger(q Quieter) *QuietLogger {
	return New(
		log.New(os.Stderr, "", log.LstdFlags),
		q)
}

// New builds a new Logger with Quieter and a Logger
func New(l Logger, q Quieter) *QuietLogger {
	return &QuietLogger{
		q: q,
		l: l,
	}
}

// Printf run logger Printf if quiet is not set
func (l *QuietLogger) Printf(format string, v ...interface{}) {
	if !l.q.Quiet() && l.l != nil {
		l.l.Printf(format, v...)
	}
}

// Fatalf run logger Fatalf if quiet is not set
func (l *QuietLogger) Fatalf(format string, v ...interface{}) {
	if !l.q.Quiet() && l.l != nil {
		l.l.Fatalf(format, v...)
	}
}
