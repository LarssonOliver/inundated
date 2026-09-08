package model

import "errors"

var (
	ErrNotFound             = errors.New("not found")
	ErrInvalidArgument      = errors.New("invalid argument")
	ErrUnprocessable        = errors.New("unprocessable")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrInvalidReference     = errors.New("invalid reference")
	ErrNotImplemented       = errors.New("not implemented")
	ErrAlreadyExists        = errors.New("already exists")
	ErrLoginStateExpired    = errors.New("login state expired")
	ErrAmbiguousOwnerScope  = errors.New("owner scope: user in context has no id")
	ErrRegistrationDisabled = errors.New("user registration is disabled")
)
