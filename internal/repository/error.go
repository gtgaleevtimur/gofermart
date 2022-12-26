package repository

import "errors"

var (
	ErrLoginAlreadyTaken  = errors.New("login already taken")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidPair        = errors.New("invalid pair: login/password")
	ErrUnauthorizedAccess = errors.New("unauthorized access detected: incident will be reported")
	ErrSessionNotFound    = errors.New("session not found")

	ErrOrderAlreadyLoadedByUser        = errors.New("the order number has already been uploaded by this user")
	ErrOrderAlreadyLoadedByAnotherUser = errors.New("the order number has already been uploaded by another user")
	ErrOrderInvalidFormat              = errors.New("invalid order number format")

	ErrTooManyRequests = errors.New("too many requests")
	ErrNoContent       = errors.New("no content")

	ErrNotEnoughFunds = errors.New("not enough funds on account")
)
