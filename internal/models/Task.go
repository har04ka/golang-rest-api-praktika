package models

import (
	"time"
)

type Task struct {
	Id          int64
	Title       string
	Description string
	CreatedAt   time.Time
	IsCompleted bool
}

type TaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	UserIDs     []int  `json:"user_ids"`
}
