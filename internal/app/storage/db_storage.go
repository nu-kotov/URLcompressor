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
			short_url      TEXT NOT NULL PRIMARY KEY,
			original_url   TEXT NOT NULL,
			correlation_id TEXT
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

func (pg *DBStorage) InsertURLsDataBatch(data []models.URLsData) error {
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}

	for _, d := range data {
		_, err := tx.ExecContext(
			context.Background(),
			`INSERT INTO urls (short_url, original_url, correlation_id) VALUES ($1, $2, $3);`,
			d.ShortURL,
			d.OriginalURL,
			d.CorrelationID,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
