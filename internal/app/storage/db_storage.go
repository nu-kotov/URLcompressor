package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nu-kotov/URLcompressor/internal/app/models"
)

type DBStorage struct {
	db *sql.DB
}

var ErrConflict = errors.New("data conflict")

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

func (pg *DBStorage) Close() error {
	return pg.db.Close()
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

func (pg *DBStorage) InsertURLsData(ctx context.Context, data *models.URLsData) error {
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO urls (short_url, original_url) VALUES ($1, $2);`,
		data.ShortURL,
		data.OriginalURL,
	)

	if err != nil {
		tx.Rollback()

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return ErrConflict
		}

		return err
	}

	return tx.Commit()
}

func (pg *DBStorage) InsertURLsDataBatch(ctx context.Context, data []models.URLsData) error {
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}

	for _, d := range data {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO urls (short_url, original_url, correlation_id) VALUES ($1, $2, $3);`,
			d.ShortURL,
			d.OriginalURL,
			d.CorrelationID,
		)
		if err != nil {
			tx.Rollback()

			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return ErrConflict
			}

			return err
		}
	}

	return tx.Commit()
}

func (pg *DBStorage) SelectOriginalURLByShortURL(ctx context.Context, shortURL string) (string, error) {
	var originalURL string

	row := pg.db.QueryRowContext(
		ctx,
		`SELECT original_url from urls WHERE short_url = $1`,
		shortURL,
	)

	err := row.Scan(&originalURL)
	if err != nil {
		return "", err
	}

	return originalURL, nil
}
