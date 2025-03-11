package repository

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// CreateUser вставляет новую запись в таблицу Users
func (db *DB) CreateUser(username, passwordHash string) (int, error) {
	var id int
	err := db.p.QueryRow(db.ctx,
		`INSERT INTO Users (username, password_hash) 
		 VALUES ($1, $2) 
		 RETURNING id;`,
		username, passwordHash,
	).Scan(&id)

	if err != nil {
		// Проверяем ошибку на дубликат (уникальный username)
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == ERRDUPLICATE {
			return 0, fmt.Errorf("user with such username %w", ErrDuplicate)
		}
		return 0, err
	}

	return id, nil
}

// GetUser возвращает пользователя по ID
func (db *DB) GetUser(username string) (*User, error) {
	var u User
	err := db.p.QueryRow(db.ctx,
		`SELECT id, username, password_hash, created_at
		 FROM Users
		 WHERE username = $1;`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ListUsers возвращает всех пользователей
func (db *DB) ListUsers() ([]User, error) {
	rows, err := db.p.Query(db.ctx,
		`SELECT id, username, password_hash, created_at FROM Users;`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// UpdateUser обновляет поля пользователя
func (db *DB) UpdateUser(userID int, username, passwordHash string) error {
	_, err := db.p.Exec(db.ctx,
		`UPDATE Users
		 SET username = $1,
		     password_hash = $2
		 WHERE id = $3;`,
		username, passwordHash, userID)
	if err != nil {
		// Проверяем ошибку на дубликат username
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == ERRDUPLICATE {
			return fmt.Errorf("user with such username %w", ErrDuplicate)
		}
		return err
	}
	return nil
}

// DeleteUser удаляет пользователя
func (db *DB) DeleteUser(userID int) error {
	_, err := db.p.Exec(db.ctx,
		`DELETE FROM Users WHERE id = $1;`,
		userID,
	)
	return err
}

// CreateService вставляет новую запись в таблицу Services
func (db *DB) CreateService(userID int, serviceName string) (int, error) {
	var id int
	err := db.p.QueryRow(db.ctx,
		`INSERT INTO Services (user_id, service_name)
		 VALUES ($1, $2)
		 RETURNING id;`,
		userID, serviceName,
	).Scan(&id)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == ERRDUPLICATE {
			return 0, fmt.Errorf("service with such name %w for this user", ErrDuplicate)
		}
		return 0, err
	}
	return id, nil
}

// GetService возвращает сервис по ID
func (db *DB) GetService(serviceID int) (*Service, error) {
	var s Service
	err := db.p.QueryRow(db.ctx,
		`SELECT id, user_id, service_name, created_at
		 FROM Services
		 WHERE id = $1;`,
		serviceID,
	).Scan(&s.ID, &s.UserID, &s.ServiceName, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ListServices возвращает все сервисы (можно фильтровать по userID)
func (db *DB) ListServices(userID int) ([]Service, error) {
	rows, err := db.p.Query(db.ctx,
		`SELECT id, user_id, service_name, created_at 
		 FROM Services
		 WHERE user_id = $1;`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var s Service
		if err := rows.Scan(&s.ID, &s.UserID, &s.ServiceName, &s.CreatedAt); err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, rows.Err()
}

// UpdateService обновляет поля сервиса
func (db *DB) UpdateService(serviceID int, serviceName string) error {
	_, err := db.p.Exec(db.ctx,
		`UPDATE Services
		 SET service_name = $1
		 WHERE id = $2;`,
		serviceName, serviceID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == ERRDUPLICATE {
			return fmt.Errorf("service with such name %w for this user", ErrDuplicate)
		}
		return err
	}
	return nil
}

// DeleteService удаляет сервис
func (db *DB) DeleteService(serviceID int) error {
	_, err := db.p.Exec(db.ctx,
		`DELETE FROM Services 
		 WHERE id = $1;`,
		serviceID,
	)
	return err
}

// CreateServiceCred вставляет новую запись в таблицу ServicesCreds
func (db *DB) CreateServiceCred(userID, serviceID int, login string, passwordEncrypt []byte, meta []byte) (int, error) {
	var id int
	err := db.p.QueryRow(db.ctx,
		`INSERT INTO ServicesCreds (user_id, service_id, login, password_encrypt, meta)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id;`,
		userID, serviceID, login, passwordEncrypt, meta,
	).Scan(&id)
	return id, err
}

// GetServiceCred возвращает запись из ServicesCreds по ID
func (db *DB) GetServiceCred(credID int) (*ServiceCred, error) {
	var sc ServiceCred
	err := db.p.QueryRow(db.ctx,
		`SELECT id, user_id, service_id, login, password_encrypt, meta, created_at, updated_at
		 FROM ServicesCreds
		 WHERE id = $1;`,
		credID,
	).Scan(
		&sc.ID, &sc.UserID, &sc.ServiceID,
		&sc.Login, &sc.PasswordEncrypt, &sc.Meta,
		&sc.CreatedAt, &sc.UpdatedAt,
	) 
	if err != nil {
		return nil, err
	}
	return &sc, nil
}

// ListServiceCreds возвращает все записи из ServicesCreds (по userID и serviceID)
func (db *DB) ListServiceCreds(userID, serviceID int) ([]ServiceCred, error) {
	rows, err := db.p.Query(db.ctx,
		`SELECT id, user_id, service_id, login, password_encrypt, meta, created_at, updated_at
		 FROM ServicesCreds
		 WHERE user_id = $1 AND service_id = $2;`,
		userID, serviceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []ServiceCred
	for rows.Next() {
		var sc ServiceCred
		if err := rows.Scan(
			&sc.ID, &sc.UserID, &sc.ServiceID,
			&sc.Login, &sc.PasswordEncrypt, &sc.Meta,
			&sc.CreatedAt, &sc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		creds = append(creds, sc)
	}
	return creds, rows.Err()
}

// UpdateServiceCred обновляет запись в ServicesCreds
func (db *DB) UpdateServiceCred(credID int, login string, passwordEncrypt, meta []byte) error {
	now := time.Now()
	_, err := db.p.Exec(db.ctx,
		`UPDATE ServicesCreds
		 SET login = $1,
		     password_encrypt = $2,
		     meta = $3,
		     updated_at = $4
		 WHERE id = $5;`,
		login, passwordEncrypt, meta, now, credID)
	return err
}

// DeleteServiceCred удаляет запись из ServicesCreds
func (db *DB) DeleteServiceCred(credID int) error {
	_, err := db.p.Exec(db.ctx,
		`DELETE FROM ServicesCreds WHERE id = $1;`,
		credID,
	)
	return err
}

// CreateTextData вставляет новую запись в таблицу TextData
func (db *DB) CreateTextData(userID, serviceID int, textData string, meta []byte) (int, error) {
	var id int
	err := db.p.QueryRow(db.ctx,
		`INSERT INTO TextData (user_id, service_id, text_data, meta)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id;`,
		userID, serviceID, textData, meta,
	).Scan(&id)
	return id, err
}

// GetTextData возвращает запись из TextData по ID
func (db *DB) GetTextData(textDataID int) (*TextData, error) {
	var td TextData
	err := db.p.QueryRow(db.ctx,
		`SELECT id, user_id, service_id, text_data, meta, created_at, updated_at
		 FROM TextData
		 WHERE id = $1;`,
		textDataID,
	).Scan(
		&td.ID, &td.UserID, &td.ServiceID,
		&td.TextData, &td.Meta, &td.CreatedAt, &td.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &td, nil
}

// ListTextData возвращает все записи из TextData (по userID и serviceID)
func (db *DB) ListTextData(userID, serviceID int) ([]TextData, error) {
	rows, err := db.p.Query(db.ctx,
		`SELECT id, user_id, service_id, text_data, meta, created_at, updated_at
		 FROM TextData
		 WHERE user_id = $1 AND service_id = $2;`,
		userID, serviceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TextData
	for rows.Next() {
		var td TextData
		if err := rows.Scan(
			&td.ID, &td.UserID, &td.ServiceID,
			&td.TextData, &td.Meta, &td.CreatedAt, &td.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, td)
	}
	return result, rows.Err()
}

// UpdateTextData обновляет запись в TextData
func (db *DB) UpdateTextData(textDataID int, textData string, meta []byte) error {
	now := time.Now()
	_, err := db.p.Exec(db.ctx,
		`UPDATE TextData
		 SET text_data = $1,
		     meta = $2,
		     updated_at = $3
		 WHERE id = $4;`,
		textData, meta, now, textDataID)
	return err
}

// DeleteTextData удаляет запись из TextData
func (db *DB) DeleteTextData(textDataID int) error {
	_, err := db.p.Exec(db.ctx,
		`DELETE FROM TextData WHERE id = $1;`,
		textDataID,
	)
	return err
}

// CreateTextBytes вставляет новую запись в таблицу TextBytes
func (db *DB) CreateTextBytes(userID, serviceID int, textBytes []byte, meta []byte) (int, error) {
	var id int
	err := db.p.QueryRow(db.ctx,
		`INSERT INTO TextBytes (user_id, service_id, text_bytes, meta)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id;`,
		userID, serviceID, textBytes, meta,
	).Scan(&id)
	return id, err
}

// GetTextBytes возвращает запись из TextBytes по ID
func (db *DB) GetTextBytes(textBytesID int) (*TextBytes, error) {
	var tb TextBytes
	err := db.p.QueryRow(db.ctx,
		`SELECT id, user_id, service_id, text_bytes, meta, created_at, updated_at
		 FROM TextBytes
		 WHERE id = $1;`,
		textBytesID,
	).Scan(
		&tb.ID, &tb.UserID, &tb.ServiceID,
		&tb.TextBytes, &tb.Meta, &tb.CreatedAt, &tb.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &tb, nil
}

// ListTextBytes возвращает все записи из TextBytes (по userID и serviceID)
func (db *DB) ListTextBytes(userID, serviceID int) ([]TextBytes, error) {
	rows, err := db.p.Query(db.ctx,
		`SELECT id, user_id, service_id, text_bytes, meta, created_at, updated_at
		 FROM TextBytes
		 WHERE user_id = $1 AND service_id = $2;`,
		userID, serviceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TextBytes
	for rows.Next() {
		var tb TextBytes
		if err := rows.Scan(
			&tb.ID, &tb.UserID, &tb.ServiceID,
			&tb.TextBytes, &tb.Meta, &tb.CreatedAt, &tb.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, tb)
	}
	return result, rows.Err()
}

// UpdateTextBytes обновляет запись в TextBytes
func (db *DB) UpdateTextBytes(textBytesID int, textBytes []byte, meta []byte) error {
	now := time.Now()
	_, err := db.p.Exec(db.ctx,
		`UPDATE TextBytes
		 SET text_bytes = $1,
		     meta = $2,
		     updated_at = $3
		 WHERE id = $4;`,
		textBytes, meta, now, textBytesID)
	return err
}

// DeleteTextBytes удаляет запись из TextBytes
func (db *DB) DeleteTextBytes(textBytesID int) error {
	_, err := db.p.Exec(db.ctx,
		`DELETE FROM TextBytes WHERE id = $1;`,
		textBytesID,
	)
	return err
}

// CreateCard вставляет новую запись в таблицу Cards
func (db *DB) CreateCard(userID, serviceID int, cardEncrypt []byte, cardLast string, expMonth, expYear int, meta []byte) (int, error) {
	var id int
	err := db.p.QueryRow(db.ctx,
		`INSERT INTO Cards (user_id, service_id, card_encrypt, card_last, exp_month, exp_year, meta)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id;`,
		userID, serviceID, cardEncrypt, cardLast, expMonth, expYear, meta,
	).Scan(&id)
	return id, err
}

// GetCard возвращает запись из Cards по ID
func (db *DB) GetCard(cardID int) (*Card, error) {
	var c Card
	err := db.p.QueryRow(db.ctx,
		`SELECT id, user_id, service_id, card_encrypt, card_last, exp_month, exp_year, meta, created_at, updated_at
		 FROM Cards
		 WHERE id = $1;`,
		cardID,
	).Scan(
		&c.ID, &c.UserID, &c.ServiceID, &c.CardEncrypt, &c.CardLast,
		&c.ExpMonth, &c.ExpYear, &c.Meta, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ListCards возвращает все записи из Cards (по userID и serviceID)
func (db *DB) ListCards(userID, serviceID int) ([]Card, error) {
	rows, err := db.p.Query(db.ctx,
		`SELECT id, user_id, service_id, card_encrypt, card_last, exp_month, exp_year, meta, created_at, updated_at
		 FROM Cards
		 WHERE user_id = $1 AND service_id = $2;`,
		userID, serviceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []Card
	for rows.Next() {
		var c Card
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.ServiceID, &c.CardEncrypt, &c.CardLast,
			&c.ExpMonth, &c.ExpYear, &c.Meta, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

// UpdateCard обновляет запись в Cards
func (db *DB) UpdateCard(cardID int, cardEncrypt []byte, cardLast string, expMonth, expYear int, meta []byte) error {
	now := time.Now()
	_, err := db.p.Exec(db.ctx,
		`UPDATE Cards
		 SET card_encrypt = $1,
		     card_last = $2,
		     exp_month = $3,
		     exp_year = $4,
		     meta = $5,
		     updated_at = $6
		 WHERE id = $7;`,
		cardEncrypt, cardLast, expMonth, expYear, meta, now, cardID,
	)
	return err
}

// DeleteCard удаляет запись из Cards
func (db *DB) DeleteCard(cardID int) error {
	_, err := db.p.Exec(db.ctx,
		`DELETE FROM Cards WHERE id = $1;`,
		cardID,
	)
	return err
}