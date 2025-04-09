package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Grifonhard/YandexGraduationProject/internal/interfaces/data"
	"github.com/Grifonhard/YandexGraduationProject/internal/logger"
	"github.com/jackc/pgx/v5"
)

const SYNC_NEW_DEVICE_TIMEOUT = 20*time.Minute

const (
	SYNC_STATUS_NO_SESS = iota			// нет сессий синхронизации
	SYNC_STATUS_ACTIVE_SES				// работает активная сессия синхронизации 
	SYNC_STATUS_CHANGE_TO_ACTIVE_SES	// запущена сессия синхронизации
)

// syncClients для синхронизации клиентов
type syncClients struct {
	sessions map[string][]*syncSession // map[имя клиента]слайс сессий с новыми клиентами
	mu *sync.Mutex
}

// syncSession отдельные сессии для синхронизации разных клиентов
type syncSession struct {
	processed bool // означает, что эта сессия синхронизации сейчас в работе
	fromNewClient *data.ClientInfo // сюда кладём инфу от нового клиента
	toNewClient chan *data.ClientData // сюда кладём инфу для нового клиента
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

	newSes := newSyncSession(info)

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

// lookForSyncSes функция для auntithicate
// если нет сессий синхронизации возвращает статус SYNC_STATUS_NO_SESS, nil
// если активная сессия уже в работе возвращает статус SYNC_STATUS_ACTIVE_SES
// если есть сессии и нет активных - переводит одну из сессий в активные, возвращает SYNC_STATUS_CHANGE_TO_ACTIVE_SES, ErrStartSync
func (s *syncClients) lookForSyncSes(username string) (status int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[username]
	if !ok {
		return SYNC_STATUS_NO_SESS, nil
	}

	if len(sess) == 0 {
		return SYNC_STATUS_NO_SESS, nil
	}

	for i := range sess {
		if sess[i].isActive() {
			return SYNC_STATUS_ACTIVE_SES, nil
		}
	}

	// если нет активных сессий, а сессии есть, то первую попавшуюсь устанавливаем в active
	sess[0].toActive()

	return SYNC_STATUS_CHANGE_TO_ACTIVE_SES, ErrStartSync	
}

// getDataForCurrentClientFromSess получаем данные для верификации нового клиента
func (s *syncClients) getDataForCurrentClientFromSess(username string) (*data.ClientInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[username]
	if !ok {
		return nil, fmt.Errorf("unexpected, sessions %w", ErrNotFound)
	}

	dataCl, err := s.getDataFromActiveSess(sess)
	if errors.Is(err, ErrUnexpActiveZero) || errors.Is(err, ErrUnexpActiveTooMuch) {
		logger.Error("%s", err.Error())
		return nil, fmt.Errorf("get data from active session: %w", err)
	} else if err != nil {
		return nil, fmt.Errorf("get data from active session: %w", err)
	}

	return dataCl, nil
}

// getDataFromActiveSess получаем данные о новом клиенте из активной сессии
// использовать только в функциях s *syncSession защищённых мьютексом
func (s *syncClients) getDataFromActiveSess(sess []*syncSession) (*data.ClientInfo, error) {
	var numActive int
	var dataCl *data.ClientInfo
	for i := range sess {
		if sess[i].isActive() {
			numActive++
			dataCl = sess[i].fromNewClient
		}
	}

	if numActive == 0 {
		return nil, ErrUnexpActiveZero
	} else if numActive > 1 {
		return nil, ErrUnexpActiveTooMuch
	}

	return dataCl, nil
}

// getDataFromSess получаем данные для нового клиента из сессии, закрываем канал и удаляем сессию после этого
func (s *syncClients) getDataForNewClientFromSess(ctx context.Context, username, tmpID string) (*data.ClientData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[username]
	if !ok {
		return nil, fmt.Errorf("unexpected, sessions %w", ErrNotFound)
	}

	for i := range sess {
		if sess[i].isMatch(tmpID) {
			select {
			case data := <-sess[i].toNewClient:

				// удаляем сессию
				newSess := append(sess[:i], sess[i+1:]...)
				close(sess[i].toNewClient)
				s.sessions[username] = newSess

				return data, nil
			case <- ctx.Done():
				return nil, ErrTimeOut
			}
		}
	}

	return nil, fmt.Errorf("unexpected, session %w", ErrNotFound)
}

// putDataToSess кладём данные для нового клиента в сессию
func (s *syncClients) putDataToSess(username, tmpID string, dataCl *data.ClientData) error {
	sess, ok := s.sessions[username]
	if !ok {
		return fmt.Errorf("unexpected, sessions %w", ErrNotFound)
	}

	for i := range sess {
		if sess[i].isMatch(tmpID) {
				if len(sess[i].toNewClient) > 0 {
					return ErrChanFullUnexp
				}
				sess[i].toNewClient <- dataCl
				return nil
		}
	}

	return fmt.Errorf("unexpected, session %w", ErrNotFound)
}

// newSyncSession новая сессия для синхронизации
// тут вкладывается информация от нового клиента
func newSyncSession(info *data.ClientInfo) *syncSession {
	var s syncSession
	s.fromNewClient = info
	s.toNewClient = make(chan *data.ClientData)
	s.mu = &sync.Mutex{}
	return &s
}

// isMatch проверяет совпадает ли id с переданным в аргументах
func (s *syncSession) isMatch(tmpID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.fromNewClient.TempID == tmpID
}

// isActive проверяет активна ли сессия
func (s *syncSession) isActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.processed
}

// toActive переводит сессию в активную
func (s *syncSession) toActive() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.processed = true
}

// SyncNewDeviceNewClient синхронизация нового клиента
// есть таймаут
func (s *Service) SyncNewDeviceClient(info *data.ClientInfo) (*data.ClientData, error) {
	_, err := s.db.GetUser(info.Uname)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, fmt.Errorf("get user from db error: %w", err)
	}

	err = s.synClients.addNewSess(info)
	if err != nil {
		return nil, err
	}

	ctx, _ := context.WithTimeout(context.Background(), SYNC_NEW_DEVICE_TIMEOUT)

	dataCl, err := s.synClients.getDataForNewClientFromSess(ctx, info.Uname, info.TempID)
	if err != nil {
		logger.Error("get data from session error: %s", err.Error())
		return nil, fmt.Errorf("get data from session error: %w", err)
	}

	if dataCl.Err != nil {
		return nil, dataCl.Err
	}

	return dataCl, nil
}

// SyncGetNewDeviceData получение данных о клиенте, чтобы проверить пароль
func (s *Service) SyncGetNewDeviceData(username string) (*data.ClientInfo, error) {
	return s.synClients.getDataForCurrentClientFromSess(username)
}

// SyncPutVerdict результат проверки нового клиента действующим
func (s *Service) SyncPutVerdict(resultMessage, username, tmpID string, dataCl *data.ClientData) error {
	var resultData data.ClientData

	// если какие-то проблемы - вкладываем ошибку, если всё нормально - данные
	switch resultMessage {
	case data.WRONG_PASS:
		resultData.Err = fmt.Errorf(data.WRONG_PASS)
	case data.NOT_APPROVED:
		resultData.Err = fmt.Errorf(data.NOT_APPROVED)
	case data.APPROVED:
		resultData.SaltPW = dataCl.SaltPW
	default:
		resultData.Err = ErrUnexpMes
	}

	return s.synClients.putDataToSess(username, tmpID, &resultData)
}
