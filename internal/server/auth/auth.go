package auth

import (
	"errors"
	"fmt"

	"github.com/Grifonhard/YandexGraduationProject/internal/server/repository"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	db *repository.DB
}

func New(db *repository.DB) (*Service, error) {
	return &Service{
		db: db,
	}, nil
}

func (s *Service) Login(username, hashpw string) (token string, err error) {
	user, err := s.db.GetUser(username)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUserNotFound
	} else if err != nil {
		return "", fmt.Errorf("get user from db error: %w", err)
	}

	if user.PasswordHash != hashpw {
		return "", ErrWrongPW
	}

	tokenDet, err := createToken(uint64(user.ID), username, EXPIRED_AT)
	if err != nil {
		return "", fmt.Errorf("create token error: %w", err)
	}

	return tokenDet.AccessToken, nil
}