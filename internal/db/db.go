package db

import (
	"easy-check/internal/config"
	"easy-check/internal/utils"
	"encoding/json"
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

func (d *DB) SaveHosts(hosts []config.Host) error {
	data, err := json.Marshal(hosts)
	if err != nil {
		return err
	}
	return d.Set("hosts", string(data))
}

// QueryStatusForHosts 根据主机列表查询对应的 StatusType
func (d *DB) QueryStatusForHosts(hosts []string) (map[string]StatusType, error) {
	statusMap := make(map[string]StatusType)

	err := d.Instance.View(func(txn *badger.Txn) error {
		for _, host := range hosts {
			key := GenerateAlertStatusKey(host)
			item, err := txn.Get(key)
			if err != nil {
				if err == badger.ErrKeyNotFound {
					// 如果主机不存在，默认状态为 RECOVERY
					statusMap[host] = StatusRecovery
					fmt.Printf("Host %s not found in DB, defaulting to RECOVERY\n", host)
					continue
				}
				return fmt.Errorf("failed to get status for host %s: %w", host, err)
			}

			var status AlertStatus
			err = item.Value(func(v []byte) error {
				return json.Unmarshal(v, &status)
			})
			if err != nil {
				return fmt.Errorf("failed to unmarshal status for host %s: %w", host, err)
			}

			// 将状态存入结果映射
			statusMap[host] = status.Status
			fmt.Printf("Host %s not found in DB, defaulting to RECOVERY\n", host)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query statuses for hosts: %w", err)
	}

	return statusMap, nil
}

// GenerateAlertStatusKey 生成主机对应的键
func GenerateAlertStatusKey(host string) []byte {
	return []byte(fmt.Sprintf("alert_status:%s", host))
}
