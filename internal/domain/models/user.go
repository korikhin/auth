package models

type User struct {
	ID           uint64 `json:"id"`
	Email        string `json:"email"`
	PasswordHash []byte `json:"-"`
	Role         string `json:"role"`
}
