package service

import "errors"

var (
	ErrLoginAlreadyTaken         = errors.New("login already taken")
	ErrUserAddedTransaction      = errors.New("add user transaction failed")
	ErrUserGetTransaction        = errors.New("get user transaction failed")
	ErrSessionGetTransaction     = errors.New("get session transaction failed")
	ErrSessionDeleteTransaction  = errors.New("delete session transaction failed")
	ErrSessionAddTransaction     = errors.New("add session transaction failed")
	ErrTypeNotAllowed            = errors.New("type not allowed")
	ErrUserNotFound              = errors.New("user not found")
	ErrInvalidPair               = errors.New("invalid pair: login/password")
	ErrSessionExistAlready       = errors.New("session already exist")
	ErrSessionNotFound           = errors.New("session not found")
	ErrOrderNotFound             = errors.New("order not found")
	ErrOrderGetTransaction       = errors.New("failed to get order")
	ErrOrderAddTransaction       = errors.New("add order transaction failed")
	ErrOrderGetTransactions      = errors.New("get all order transaction failed")
	ErrBalanceGetTransaction     = errors.New("get balance transaction failed")
	ErrWithdrawalPostTransaction = errors.New("post withdraw transaction failed")
	ErrWithdrawalsGetTransaction = errors.New("get withdrawals transaction failed")

	ErrOrderAlreadyLoadedByUser        = errors.New("the order number has already been uploaded by this user")
	ErrOrderAlreadyLoadedByAnotherUser = errors.New("the order number has already been uploaded by another user")
	ErrOrderInvalidFormat              = errors.New("invalid order number format")

	ErrTooManyRequests = errors.New("too many requests")
	ErrNoContent       = errors.New("no content")

	ErrNotEnoughFunds = errors.New("not enough funds on account")
)
