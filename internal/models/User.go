package models

import "time"

type User struct {
	Id           int64
	Login        string
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

type UserPublicResponse struct {
	Id      int    `json:"id"`
	Family  string `json:"family"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
}

type UserProfileResponse struct {
	Id      int    `json:"id"`
	Family  string `json:"family"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	IsAdmin bool   `json:"is_admin"`
}

type LoginResponse struct {
	Status string              `json:"status"`
	Token  string              `json:"token"`
	User   UserProfileResponse `json:"user"`
}

type UserRequest struct {
	Login    string `json:"login"`
	Family   string `json:"family"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Password string `json:"password"`
}
