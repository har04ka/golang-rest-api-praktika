package models

import "time"

type Session struct {
	Id        int64
	UserId    int64
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}
