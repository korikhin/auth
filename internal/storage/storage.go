package storage

import "errors"

var (
	ErrConnectionFailed  = errors.New("failed to connect to the storage")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)
