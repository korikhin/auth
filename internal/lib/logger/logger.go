package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/korikhin/auth/internal/config"
	"github.com/korikhin/auth/internal/domain/models"
)

func New(s config.Stage) *slog.Logger {
	var h slog.Handler
	opts := &slog.HandlerOptions{ReplaceAttr: replaceAttr}

	switch s {
	case config.Local:
		opts.Level = slog.LevelDebug
		h = slog.NewTextHandler(os.Stdout, opts)
	case config.Dev:
		opts.Level = slog.LevelDebug
		h = slog.NewJSONHandler(os.Stdout, opts)
	case config.Prod:
		opts.Level = slog.LevelInfo
		h = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(h)
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if len(groups) == 0 {
		// Enforce UTC
		if a.Key == slog.TimeKey {
			a.Value = slog.TimeValue(a.Value.Time().UTC())
		}
		// Discard empty message
		if a.Equal(slog.String(slog.MessageKey, "")) {
			return slog.Attr{}
		}
		// Discard nil error
		if a.Equal(Error(nil)) {
			return slog.Attr{}
		}
	}

	return a
}

func Component(c string) slog.Attr {
	return slog.String("component", c)
}

func Duration(d time.Duration) slog.Attr {
	return slog.Duration("duration_nanos", d)
}

func Error(err error) slog.Attr {
	return slog.Any("error", err)
}

func Operation(op string) slog.Attr {
	return slog.String("operation", op)
}

func RequestID(id string) slog.Attr {
	return slog.String("request_id", id)
}

func Signal(s os.Signal) slog.Attr {
	return slog.String("signal", s.String())
}

func Stage(e config.Stage) slog.Attr {
	return slog.String("stage", string(e))
}

func User(u models.User) slog.Attr {
	return slog.Any("user", u)
}
