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
	ErrNotFound = errors.New("not found")
	ErrTimeOut = errors.New("time out")
	ErrUnexpMes = errors.New("unexpected message from current client")
	ErrChanFullUnexp = errors.New("unexpected channel is full")
	ErrChanEmptyUnexp = errors.New("unexpected channel is empty")
	ErrStartSync = errors.New("start synchronization session")
	ErrUnexpActiveTooMuch = errors.New("unexpectedly too many active sessions")
	ErrUnexpActiveZero = errors.New("unexpectedly no active sessions")
)