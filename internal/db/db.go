package db

import (
	"github.com/dgraph-io/badger/v4"
)

type DB struct {
	db *badger.DB
}

// NewDB 初始化 BadgerDB
func NewDB(path string) (*DB, error) {
	opts := badger.DefaultOptions(path).WithLoggingLevel(badger.ERROR) // 设置日志级别为 ERROR
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

// Set 设置键值对
func (d *DB) Set(key, value string) error {
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(value))
	})
}

// Get 获取键对应的值
func (d *DB) Get(key string) (string, error) {
	var val string
	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {
			val = string(v)
			return nil
		})
	})
	if err != nil {
		return "", err
	}
	return val, nil
}

// Delete 删除键
func (d *DB) Delete(key string) error {
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// Close 关闭数据库
func (d *DB) Close() error {
	return d.db.Close()
}
