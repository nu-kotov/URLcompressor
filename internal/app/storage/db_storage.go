package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nu-kotov/URLcompressor/internal/app/models"
	"github.com/pressly/goose/v3"
)

type DBStorage struct {
	db *sql.DB
}

var ErrConflict = errors.New("data conflict")

var (
	dbInstance *DBStorage
	//go:embed migrations/*.sql
	embedMigrations embed.FS
)

func NewConnect(connString string) (*DBStorage, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}

	if err := goose.Up(db, "migrations"); err != nil {
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

func (pg *DBStorage) InsertURLsData(ctx context.Context, data *models.URLsData) error {

	sql := `INSERT INTO urls (short_url, original_url) VALUES ($1, $2);`

	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		sql,
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

	sql := `INSERT INTO urls (short_url, original_url, correlation_id) VALUES ($1, $2, $3);`

	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}

	for _, d := range data {
		_, err := tx.ExecContext(
			ctx,
			sql,
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

	sql := `SELECT original_url from urls WHERE short_url = $1`

	row := pg.db.QueryRowContext(
		ctx,
		sql,
		shortURL,
	)

	err := row.Scan(&originalURL)
	if err != nil {
		return "", err
	}

	return originalURL, nil
}
