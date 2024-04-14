package models

import "log/slog"

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash []byte `json:"-"`
	// Role         string `json:"role"`
}

func (u *User) LogValue() slog.Value {
	return slog.StringValue(u.ID)
}
