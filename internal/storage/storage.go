package storage

import "errors"

var (
	ErrConnectionFailed  = errors.New("failed to connect to the storage")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)
