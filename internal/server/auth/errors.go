package auth

import "errors"

var (
	ErrUnexpSignMethod = errors.New("unexpected signing method")
	ErrInvalidMethod = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
	ErrNoExpTime = errors.New("invalid token, no exp")
	ErrUserNotFound = errors.New("user not found")
	ErrWrongPW = errors.New("wrong password")
	ErrEmptyToken = errors.New("empty token")
	ErrAuthFail = errors.New("authentication failed")
	ErrBadToken = errors.New("bad token")
	ErrBadClaims = errors.New("bad claims")
	ErrBadClient = errors.New("the information about the mac address and the client's IP do not match")
	ErrWasUsed = errors.New("one-time password has already been used")
)

// sync
var (
	ErrAlreadyStart = errors.New("already being processed")
)