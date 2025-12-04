package utils

import (
	"strings"
)

func ValidateUserRequest(login, family, name, surname, password string) error {
	if strings.TrimSpace(login) == "" {
		return &ValidationError{Field: "login", Message: "login is required"}
	}
	if strings.TrimSpace(family) == "" {
		return &ValidationError{Field: "family", Message: "family is required"}
	}
	if strings.TrimSpace(name) == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	if strings.TrimSpace(surname) == "" {
		return &ValidationError{Field: "surname", Message: "surname is required"}
	}
	if strings.TrimSpace(password) == "" {
		return &ValidationError{Field: "password", Message: "password is required"}
	}
	if len(password) < 6 {
		return &ValidationError{Field: "password", Message: "password must be at least 6 characters"}
	}
	return nil
}

func ValidateTaskRequest(title, description string) error {
	if strings.TrimSpace(title) == "" {
		return &ValidationError{Field: "title", Message: "title is required"}
	}
	return nil
}

func ValidateLoginRequest(login, password string) error {
	if strings.TrimSpace(login) == "" {
		return &ValidationError{Field: "login", Message: "login is required"}
	}
	if strings.TrimSpace(password) == "" {
		return &ValidationError{Field: "password", Message: "password is required"}
	}
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

