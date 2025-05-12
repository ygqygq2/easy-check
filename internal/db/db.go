package db

import (
	"easy-check/internal/config"
	"easy-check/internal/utils"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

type DB struct {
	Instance *badger.DB
}

// NewDB 初始化 BadgerDB
func NewDB(isDev bool, dbConfig *config.DbConfig) (*DB, error) {
	badgerPath := "badger"
	var opts badger.Options
	if !isDev {
		opts = badger.DefaultOptions(utils.AddDirectorySuffix(dbConfig.Path) + badgerPath).WithLoggingLevel(badger.ERROR) // 设置日志级别为 ERROR
	} else {
		opts = badger.DefaultOptions("").WithInMemory(true).WithLoggingLevel(badger.ERROR)
	}
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open Badger DB: %v", err)
	}
	return &DB{Instance: db}, nil
}

// Set 设置键值对
func (d *DB) Set(key, value string) error {
	return d.Instance.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(value))
	})
}

// Get 获取键对应的值
func (d *DB) Get(key string) (string, error) {
	var val string
	err := d.Instance.View(func(txn *badger.Txn) error {
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
	return d.Instance.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// Close 关闭数据库
func (d *DB) Close() error {
	return d.Instance.Close()
}
