package repository

import (
	"database/sql"
	"fmt"
)

type Transaction interface {
	Begin() error
	Rollback() error
	Commit() error
	ReceiptDetectionHistoriesPostgresTx() *receiptDetectionHistoriesPostgres
}

type sqlTransaction struct {
	db *sql.DB
	tx *sql.Tx
}

func NewSqlTransaction(db *sql.DB) *sqlTransaction {
	return &sqlTransaction{
		db: db,
	}
}

func (s *sqlTransaction) Begin() error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("[transaction][Begin][db.Begin] Error: %w", err)
	}

	s.tx = tx

	return nil
}

func (s *sqlTransaction) Rollback() error {
	return s.tx.Rollback()
}

func (s *sqlTransaction) Commit() error {
	return s.tx.Commit()
}

func (s *sqlTransaction) ReceiptDetectionHistoriesTx() *receiptDetectionHistoriesPostgres {
	return &receiptDetectionHistoriesPostgres{
		dbtx: s.tx,
	}
}
