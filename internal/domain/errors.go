package domain

import "errors"

var (
	ErrNotFound            = errors.New("resource not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrInvalidInput        = errors.New("invalid input provided")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrWizardAlreadyDone   = errors.New("server is already setup and setup wizard is disabled")
	ErrSessionNotFound     = errors.New("streaming session not found")
	ErrSessionAlreadyEnded = errors.New("streaming session has already ended")
)
