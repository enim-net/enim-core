package sentry

import (
	"context"
	"errors"
	"time"
	
	sentrygo "github.com/getsentry/sentry-go"
	
	"github.com/enim-net/enim-core/logger"
)

type Config struct {
	DSN          string
	Environment  string
	Release      string
	MinLevel     logger.Level
	SampleRate   float64
	FlushTimeout time.Duration
	Debug        bool
}

type Adapter struct {
	hub          *sentrygo.Hub
	min          logger.Level
	flushTimeout time.Duration
}

func New(cfg Config) (*Adapter, error) {
	if cfg.DSN == "" {
		return nil, errors.New("sentry: missing sentry driver DSN")
	}
	if cfg.SampleRate == 0 {
		cfg.SampleRate = 1.0
	}
	if cfg.FlushTimeout <= 0 {
		cfg.FlushTimeout = 3 * time.Second
	}
	if cfg.MinLevel == logger.LevelDebug {
		cfg.MinLevel = logger.LevelDebug
	}
	
	client, err := sentrygo.NewClient(sentrygo.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		SampleRate:       cfg.SampleRate,
		AttachStacktrace: true,
		Debug:            cfg.Debug,
	})
	
	if err != nil {
		return nil, err
	}
	return &Adapter{
		hub:          sentrygo.NewHub(client, sentrygo.NewScope()),
		min:          cfg.MinLevel,
		flushTimeout: cfg.FlushTimeout,
	}, nil
}

func (a *Adapter) Write(_ context.Context, e logger.Entry) error {
	if e.Level < a.min {
		return nil
	}
	
	a.hub.WithScope(func(scope *sentrygo.Scope) {
		scope.SetLevel(toSentryLevel(e.Level))
		
		fields := sentrygo.Context{}
		for _, f := range e.Fields {
			fields[f.Key] = f.Value
		}
		if len(fields) > 0 {
			scope.SetContext("fields", fields)
		}
		if e.Caller != "" {
			scope.SetTag("caller", e.Caller)
		}
		
		if e.Error != nil {
			// Keep the log message visible next to the exception.
			scope.SetContext("log", sentrygo.Context{"message": e.Message})
			a.hub.CaptureException(e.Error)
			return
		}
		a.hub.CaptureMessage(e.Message)
	})
	return nil
}

// Close flushes queued events, waiting up to ctx's deadline (or FlushTimeout).
func (a *Adapter) Close(ctx context.Context) error {
	timeout := a.flushTimeout
	if dl, ok := ctx.Deadline(); ok {
		timeout = time.Until(dl)
	}
	if !a.hub.Flush(timeout) {
		return errors.New("sentry: flush timed out, some events may be lost")
	}
	return nil
}

func toSentryLevel(l logger.Level) sentrygo.Level {
	switch l {
	case logger.LevelDebug:
		return sentrygo.LevelDebug
	case logger.LevelInfo:
		return sentrygo.LevelInfo
	case logger.LevelWarn:
		return sentrygo.LevelWarning
	default:
		return sentrygo.LevelError
	}
}
