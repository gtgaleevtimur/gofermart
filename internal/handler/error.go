package handler

import "errors"

var (
	ErrUnauthorizedAccess = errors.New("unauthorized access detected: incident will be reported")
	ErrSessionExpired     = errors.New("session has expired")
)
