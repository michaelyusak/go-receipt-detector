package repository

import (
	"database/sql"
	"fmt"
)

type sqlTransaction struct {
	db *sql.DB
	tx *sql.Tx
}

func NewSqlTransaction(db *sql.DB) *sqlTransaction {
	return &sqlTransaction{
		db: db,
	}
}

func (s *sqlTransaction) Begin() (*sql.Tx, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("[transaction][Begin][db.Begin] Error: %w", err)
	}

	return tx, nil
}

func (s *sqlTransaction) Rollback() error {
	return s.tx.Rollback()
}

func (s *sqlTransaction) Commit() error {
	return s.tx.Commit()
}
