package auth

import "errors"

var (
	ErrUnexpSignMethod = errors.New("unexpected signing method")
	ErrInvalidMethod = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
	ErrNoExpTime = errors.New("invalid token, no exp")
	ErrUserNotFound = errors.New("user not found")
	ErrWrongPW = errors.New("wrong password")
)