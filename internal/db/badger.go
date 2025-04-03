package db

import (
	"github.com/dgraph-io/badger/v4"
)

func (d *DB) SetAlertStatus(host, status string) error {
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(host), []byte(status))
	})
}

func (d *DB) GetAlertStatus(host string) (string, error) {
	var status string
	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(host))
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {
			status = string(v)
			return nil
		})
	})
	if err != nil {
		return "", err
	}
	return status, nil
}

func (d *DB) DeleteAlertStatus(host string) error {
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(host))
	})
}

func (d *DB) Close() error {
	return d.db.Close()
}
