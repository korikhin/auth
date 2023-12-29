package models

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash []byte `json:"-"`
	// Role         string `json:"role"`
}
