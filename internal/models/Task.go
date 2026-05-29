package models

import (
	"time"
)

type Task struct {
	Id          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	IsCompleted bool      `json:"is_completed"`
}

type TaskRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Is_completed bool   `json:"is_completed"`
}
