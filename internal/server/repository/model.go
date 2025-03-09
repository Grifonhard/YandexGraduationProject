package repository

import "time"

type User struct {
	ID           int
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type Service struct {
	ID          int
	UserID      int
	ServiceName string
	CreatedAt   time.Time
}

type ServiceCred struct {
	ID              int
	UserID          int
	ServiceID       int
	Login           string
	PasswordEncrypt []byte
	Meta            []byte // JSONB
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}

type TextData struct {
	ID         int
	UserID     int
	ServiceID  int
	TextData   string
	Meta       []byte // JSONB
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}

type TextBytes struct {
	ID         int
	UserID     int
	ServiceID  int
	TextBytes  []byte
	Meta       []byte // JSONB
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}

type Card struct {
	ID          int
	UserID      int
	ServiceID   int
	CardEncrypt []byte
	CardLast    string
	ExpMonth    int
	ExpYear     int
	Meta        []byte // JSONB
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}
