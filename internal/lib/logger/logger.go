package logger

import (
	"log/slog"
	"os"

	"github.com/studopolis/auth-server/internal/config"
)

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

func New(s config.Stage) *slog.Logger {
	var h slog.Handler
	opts := &slog.HandlerOptions{
		ReplaceAttr: replaceAttr,
	}

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
