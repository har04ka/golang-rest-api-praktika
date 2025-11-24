package models

import "time"

type User struct {
	Id           int64
	Family       string
	Name         string
	Surname      string
	Password     string
	IsAdmin      bool
	SessionToken *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserResponse struct {
	Id        int       `json:"id"`
	Family    string    `json:"family"`
	Name      string    `json:"name"`
	Surname   string    `json:"surname"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRequest struct {
	Family   string `json:"family"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Password string `json:"password"`
}
