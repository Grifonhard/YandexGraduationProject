package repository

import "errors"

var (
	ErrDuplicate = errors.New("already exists")
)