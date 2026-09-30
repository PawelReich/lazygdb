package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Db struct {
	db *sql.DB

	cwd string
}

func NewDb() (*Db, error) {
	dbPath, err := getDbPath()
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_timeout=1000&_sync=NORMAL", dbPath)

	sqLiteDb, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	db := &Db{db: sqLiteDb, cwd: cwd}

	db.EnsureCommandHistoryTable()

	return db, nil
}

func getDbPath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	dbDir := filepath.Join(cacheDir, "lazygdb")
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		return "", err
	}

	dbPath := filepath.Join(dbDir, "lazygdb.db")

	return dbPath, nil
}
