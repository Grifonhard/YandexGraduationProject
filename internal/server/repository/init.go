package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	p *pgxpool.Pool
	ctx context.Context
}

const (
	ERRDUPLICATE = "23505"
)

func New(uri string) (*DB, error) {
	var db DB
	var err error
	db.p, err = pgxpool.New(context.Background(), uri)
	if err != nil {
		return nil, err
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)

	db.ctx = ctx

	err = db.CreateTables()
	if err != nil {
		return nil, err
	}

	return &db, nil
}

func (db *DB) CreateTables() error {
	// TODO связь с BalanceTransactions обновить
	_, err := db.p.Exec(db.ctx, `
		CREATE TABLE IF NOT EXISTS Users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS Services (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES Users(id) ON DELETE CASCADE,
			service_name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			-- Уникальная пара (user_id, service_name), чтобы
			-- в рамках одного user_id не было двух одинаковых service_name
			CONSTRAINT services_userid_sname_uniq UNIQUE (user_id, service_name),
			-- Составной уникальный ключ (id, user_id), чтобы можно было сослаться на них вместе
			CONSTRAINT services_id_userid_uniq UNIQUE (id, user_id)
		);

		CREATE TABLE IF NOT EXISTS ServicesCreds (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES Users(id) ON DELETE CASCADE,
			service_id INT NOT NULL REFERENCES Services(id) ON DELETE CASCADE,
			login VARCHAR(255),
			password_encrypt BYTEA NOT NULL,
			meta JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP,
			-- Составной FOREIGN KEY (service_id, user_id) -> Services (id, user_id)
			CONSTRAINT fk_service_user
				FOREIGN KEY (service_id, user_id)
				REFERENCES Services (id, user_id)
				ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS TextData (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES Users(id) ON DELETE CASCADE,
			service_id INT NOT NULL REFERENCES Services(id) ON DELETE CASCADE,
			text_data TEXT,
			meta JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP,
			CONSTRAINT fk_td_service_user
				FOREIGN KEY (service_id, user_id)
				REFERENCES Services (id, user_id)
				ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS TextBytes (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES Users(id) ON DELETE CASCADE,
			service_id INT NOT NULL REFERENCES Services(id) ON DELETE CASCADE,
			text_bytes BYTEA,
			meta JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP,
			CONSTRAINT fk_tb_service_user
				FOREIGN KEY (service_id, user_id)
				REFERENCES Services (id, user_id)
				ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS Cards (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES Users(id) ON DELETE CASCADE,
			service_id INT NOT NULL REFERENCES Services(id) ON DELETE CASCADE,
			card_encrypt BYTEA,
			card_last VARCHAR(4),
			exp_month SMALLINT,
			exp_year SMALLINT,
			meta JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP,
			CONSTRAINT fk_c_service_user
				FOREIGN KEY (service_id, user_id)
				REFERENCES Services (id, user_id)
				ON DELETE CASCADE
		);`)
	return err
}

func (db *DB) Close() {
	db.p.Close()
}