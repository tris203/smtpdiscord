package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS domains (domain TEXT PRIMARY KEY, webhook_url TEXT)")
	if err != nil {
		return nil, err
	}

	return db, nil
}
