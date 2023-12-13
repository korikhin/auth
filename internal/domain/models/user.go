package models

type User struct {
	ID           string `json:"id"`
	Nickname     string `json:"nickname,omitempty"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
}
