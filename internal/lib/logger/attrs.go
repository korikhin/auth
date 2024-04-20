package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/korikhin/auth/internal/config"
	"github.com/korikhin/auth/internal/domain/models"
)

func Component(c string) slog.Attr {
	return slog.String("component", c)
}

func Duration(d time.Duration) slog.Attr {
	return slog.Duration("duration", d)
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
