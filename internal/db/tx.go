package db

import "database/sql"

type TxFn func(*sql.Tx) error

func WithTx(db *sql.DB, fn TxFn) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
