package model

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInvalidReference = errors.New("invalid reference")
	ErrNotImplemented   = errors.New("not implemented")
)
