package storage

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nu-kotov/URLcompressor/internal/app/models"
)

type DBStorage struct {
	db *sql.DB
}

var (
	dbInstance *DBStorage
)

func NewConnect(connString string) (*DBStorage, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, err
	}
	dbInstance = &DBStorage{db}

	return dbInstance, nil
}

func (pg *DBStorage) Ping() error {
	return pg.db.Ping()
}

func (pg *DBStorage) Close() {
	pg.db.Close()
}

func (pg *DBStorage) CreateTable() error {
	_, err := pg.db.ExecContext(
		context.Background(),
		`CREATE TABLE IF NOT EXISTS urls (
			short_url    TEXT NOT NULL PRIMARY KEY,
			original_url TEXT NOT NULL
		);`)

	if err != nil {
		return err
	}
	return nil
}

func (pg *DBStorage) InsertURLsData(data *models.URLsData) error {
	_, err := pg.db.ExecContext(
		context.Background(),
		`INSERT INTO urls (short_url, original_url) VALUES ($1, $2);`,
		data.ShortURL,
		data.OriginalURL,
	)

	if err != nil {
		return err
	}
	return nil
}
