package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"url-shortener/internal/storage"

	// _ "github.com/mattn/go-sqlite3" // initializating driver for DB sqlite3
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const fn = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS url(
        id INTEGER PRIMARY KEY,
        alias TEXT NOT NULL UNIQUE,
        url TEXT NOT NULL);
CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);

	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	exec, err := stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	_ = exec

	return &Storage{db: db}, nil

}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const fn = "storage.sqlite3.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		// if add url with alias which already was in DB -> erorr
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrURLExists) //storage.ErrURLExists -> common erorr for all storages (headler will know what to do, for exmp: "this url already in DB")
		}

		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	id, err := res.LastInsertId() // get last inserted id(supported not by all DB)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const fn = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?") // preparing query statement
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", fn, err)
	}

	var resURL string

	err = stmt.QueryRow(alias).Scan(&resURL) // searching in DB
	if errors.Is(err, sql.ErrNoRows) {       // if url does not exists
		return "", storage.ErrURLNotFound
	}

	if err != nil { // if wrong with query statement
		fmt.Errorf("%s: execute statement %w", fn, err)
	}

	return resURL, nil

}

// TODO
// func (s *Storage) DeleteURL(alias string) error{}
