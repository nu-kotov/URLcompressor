package storage

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
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
