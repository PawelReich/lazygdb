package db

import (
	_ "embed"
)

//go:embed schema/command_history.sql
var schemaCommandHistory string

func (db *Db) EnsureCommandHistoryTable() error {
	_, err := db.db.Exec(schemaCommandHistory)
	return err
}

func (db *Db) InsertHistory(command string) error {
	query := "INSERT INTO command_history (cwd, command) VALUES (?, ?)"
	_, err := db.db.Exec(query, db.cwd, command)
	return err
}

func (db *Db) GetCommandFromHistory(index uint) (string, error) {
	query := "SELECT command FROM command_history WHERE cwd = ? ORDER BY id DESC LIMIT 1 OFFSET ?"
	offset := index
	// panic(offset)
	res := db.db.QueryRow(query, db.cwd, offset)

	var command string
	err := res.Scan(&command)
	return command, err
}
