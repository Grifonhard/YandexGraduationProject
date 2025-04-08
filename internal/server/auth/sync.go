package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Grifonhard/YandexGraduationProject/internal/interfaces/data"
	"github.com/jackc/pgx/v5"
)

// syncClients для синхронизации клиентов
type syncClients struct {
	sessions map[string][]*syncSession // map[имя клиента]слайс сессий с новыми клиентами
	mu *sync.Mutex
}

// syncSession отдельные сессии для синхронизации разных клиентов
type syncSession struct {
	id string // временный идентефикатор нового клиента, генерируется на стороне нового клиента
	processed bool // означает, что эта сессия синхронизации сейчас в работе
	fromNewClient chan data.ClientInfo // сюда кладём инфу от нового клиента
	toNewClient chan data.ClientInfo // сюда кладём инфу для нового клиента
	mu *sync.Mutex
}

// newSyncClients новый экземпляр мапы для синхронизации клиентов
func newSyncClients() *syncClients {
	var s syncClients
	s.sessions = make(map[string][]*syncSession)
	s.mu = &sync.Mutex{}
	return &s
}

// addNewSess создаёт новую сессию синхронизации нового клиента с действующими
func (s *syncClients) addNewSess(info *data.ClientInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isExist(info.Uname, info.TempID) {
		return fmt.Errorf("New client confirmation request is %w", ErrAlreadyStart)
	}

	newSes := newSyncSession(info.TempID)

	s.sessions[info.Uname] = append(s.sessions[info.Uname], newSes)

	return nil
}

// isExist проверяет есть ли сессия с этим айдишником в работе
// использовать только в функциях s *syncSession защищённых мьютексом
func (s *syncClients) isExist(username, tmpID string) bool {
	sess, ok := s.sessions[username]
	if !ok {
		return false
	}
	for i := range sess {
		if sess[i].isMatch(tmpID) {
			return true
		}
	}
	return false
}

// newSyncSession новая сессия для синхронизации
func newSyncSession(tempIDnewClient string) *syncSession {
	var s syncSession
	s.id = tempIDnewClient
	s.mu = &sync.Mutex{}
	return &s
}

// isMatch проверяет совпадает ли id с переданным в аргументах
func (s *syncSession) isMatch(tmpID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.id == tmpID
}



// SyncNewDeviceNewClient синхронизация нового девайса
func (s *Service) SyncNewDeviceNewClient(ctx context.Context, info *data.ClientInfo) (*data.ClientData, error) {
	_, err := s.db.GetUser(info.Uname)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, fmt.Errorf("get user from db error: %w", err)
	}

	

	

	return data.ClientData{}, nil
}

func (s *Service) GetNewDeviceInfo()
