package webring

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(databasePath string) (*Database, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?cache=shared", databasePath))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS gopherspaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			domain TEXT NOT NULL,
			port INTEGER NOT NULL,
			path TEXT NOT NULL,
			token TEXT NOT NULL,
			approved BOOLEAN DEFAULT 0,
			UNIQUE(domain, port, path),
			UNIQUE(token)
		);
			
		CREATE TABLE IF NOT EXISTS webring (
			gopherspace_id INTEGER PRIMARY KEY,
			previous INTEGER NOT NULL DEFAULT -1,
			next INTEGER NOT NULL DEFAULT -1,
			FOREIGN KEY (gopherspace_id) REFERENCES gopherspaces(id) ON DELETE CASCADE
		)`,
	)

	if err != nil {
		return nil, err
	}

	return &Database{
		db: db,
	}, nil
}
