package logger

import (
	"log/slog"
	"time"
)

func Component(c string) slog.Attr {
	return slog.Attr{
		Key:   "component",
		Value: slog.StringValue(c),
	}
}

func Duration(d time.Duration) slog.Attr {
	return slog.Attr{
		Key:   "duration",
		Value: slog.StringValue(d.String()),
	}
}

func Error(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func Operation(op string) slog.Attr {
	return slog.Attr{
		Key:   "op",
		Value: slog.StringValue(op),
	}
}

func RequestID(id string) slog.Attr {
	return slog.Attr{
		Key:   "request_id",
		Value: slog.StringValue(id),
	}
}

func User(id string) slog.Attr {
	return slog.Attr{
		Key:   "user",
		Value: slog.StringValue(id),
	}
}
