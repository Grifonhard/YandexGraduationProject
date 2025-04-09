package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/Grifonhard/YandexGraduationProject/internal/logger"
	"github.com/Grifonhard/YandexGraduationProject/internal/server/repository"
	"github.com/Grifonhard/YandexGraduationProject/internal/utils"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	db *repository.DB
	synClients *syncClients
}

func New(db *repository.DB) (*Service, error) {
	return &Service{
		db: db,
		synClients: newSyncClients(),
	}, nil
}

type User struct {
	IsAutorized bool
	TokenUUID string
	UserID int
	Username string
	ExpiredAt time.Time
}

// Login создаём токен, вносим в бд запись о созданном uuid токена
func (s *Service) Login(username, hashpw string) (token string, err error) {
	user, err := s.db.GetUser(username)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUserNotFound
	} else if err != nil {
		return "", fmt.Errorf("get user from db error: %w", err)
	}

	if user.Disposable && !user.IsActive {
		return "", ErrWasUsed
	}

	if user.PasswordHash != hashpw {
		return "", ErrWrongPW
	}

	tokenDet, err := createToken(user.ID, username, EXPIRED_AT)
	if err != nil {
		return "", fmt.Errorf("create token error: %w", err)
	}

	// добавляем в бд информацию о созданном токене
	expired := time.Unix(tokenDet.AtExpires, 0)
	_, err = s.db.CreateToken(user.ID, tokenDet.AccessUUID, &expired)
	if err != nil {
		return "", fmt.Errorf("add to db new token error: %w", err)
	}

	return tokenDet.AccessToken, nil
}

// Authenticate проверяем токен, сравниваем с записью в бд
func (s *Service) Authenticate(token, mac, ip string) (userInfo *User, err error) {
	if token == "" {
		return nil, ErrEmptyToken
	}

	t, err := decodeToken(token)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrBadToken, err.Error())
	}

	err = checkClient(token, mac, ip)
	if err != nil {
		return nil, err
	}

	userInfo, err = getUserFromClaims(t)
	if err != nil {
		return nil, err
	}

	tokInfo, err := s.db.GetToken(userInfo.TokenUUID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrBadToken, err.Error())
	}

	if tokInfo == nil {
		return nil, ErrBadToken
	}

	if tokInfo.UserID != userInfo.UserID {
		return nil, fmt.Errorf("foreign token %w", ErrBadToken)
	}

	// если необходима синхронизация другого девайса возвращает специальную ошибку
	status, err := s.synClients.lookForSyncSes(userInfo.Username)
	switch status {
	case SYNC_STATUS_CHANGE_TO_ACTIVE_SES:
		if err != nil {
			return nil, err
		} else {
			logger.Error("missing special error")
		}
	default:
		if err != nil {
			return nil, err
		}
	}

	return userInfo, nil
}

// ChangePassword меняем хэш пароля в базе
func (s *Service) ChangePassword(uname, oldPW, newPW string) error {
	user, err := s.db.GetUser(uname)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	} else if err != nil {
		return fmt.Errorf("get user from db error: %w", err)
	}

	if user.PasswordHash != oldPW {
		return ErrWrongPW
	}

	return s.db.UpdateUser(user.ID, "", newPW)
}

// ChangeUsername меняет username
func (s *Service) ChangeUsername(currentUname, newUname string) error {
	user, err := s.db.GetUser(currentUname)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	} else if err != nil {
		return fmt.Errorf("get user from db error: %w", err)
	}

	return s.db.UpdateUser(user.ID, newUname, "")
}

// Logout удаляем токен из списка действующих
func (s *Service) Logout(info *User) error {
	token, err := s.db.GetToken(info.TokenUUID)
	if err != nil {
		return fmt.Errorf("get token error: %w", err)
	}
	return s.db.DeleteToken(token.ID)
}

// CleanExpiredTokens удаляет просроченные токены в базе
func (s *Service) CleanExpiredTokens() error {
	tokens, err := s.db.ListTokens() 
	if err != nil {
		return fmt.Errorf("list token error: %w", err)
	}

	for i := range tokens {
		if tokens[i].ExpiredAt.Before(time.Now()) {
			err = s.db.DeleteToken(tokens[i].ID)
			if err != nil {
				return fmt.Errorf("delete token error: %w", err)
			}
		}
	}
	return nil
}

// CreateUser создание нового юзера
func (s *Service) CreateUser(username string) (tempPW string, err error) {
	tempPW, err = utils.GenerateString(10)
	if err != nil {
		return "", err
	}

	_, err = s.db.CreateUser(username, tempPW)
	if err != nil {
		return "", err
	}

	return tempPW, nil
}
