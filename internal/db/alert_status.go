package db

import (
	"easy-check/internal/config"
	"easy-check/internal/logger"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type StatusType string
type SentType bool

const (
	StatusAlert    StatusType = "ALERT"
	StatusRecovery StatusType = "RECOVERY"
)

const (
	SentTrue  SentType = true
	SentFalse SentType = false
)

// AlertStatusManager 提供对告警状态的存储和管理功能
type AlertStatusManager struct {
	db       *badger.DB
	logger   *logger.Logger
	dbConfig *config.DbConfig
}

type AlertStatus struct {
	Host         string     `json:"host"`
	Description  string     `json:"description"`
	FailAlert    bool       `json:"fail_alert"`
	Status       StatusType `json:"status"`
	FailTime     string     `json:"fail_time"`
	RecoveryTime string     `json:"recovery_time"`
	Sent         bool       `json:"sent"`
}

// NewAlertStatusManager 创建一个新的 AlertStatusManager
func NewAlertStatusManager(dbInstance *badger.DB, logger *logger.Logger, dbConfig config.DbConfig) (*AlertStatusManager, error) {
	if dbInstance == nil {
		return nil, logger.LogAndError("DBInstance is nil, cannot create AlertStatusManager", "error")
	}
	return &AlertStatusManager{db: dbInstance, logger: logger, dbConfig: &dbConfig}, nil
}

// generateKey 生成 Badger 键
func (d *AlertStatusManager) generateKey(host string) []byte {
	return []byte(fmt.Sprintf("alert_status:%s", host))
}

// SetAlertStatus 保存告警状态
func (d *AlertStatusManager) SetAlertStatus(status AlertStatus, ttlSeconds int) error {
	if d == nil {
		return d.logger.LogAndError("AlertStatusManager instance is nil", "error")
	}
	if d.db == nil {
		return d.logger.LogAndError("database instance is nil in AlertStatusManager", "error")
	}
	key := d.generateKey(status.Host)
	value, err := json.Marshal(status)
	if err != nil {
		return d.logger.LogAndError("failed to marshal alert status: %v", "error", err)
	}

	// 设置带有 TTL 的键值对
	return d.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, value).WithTTL(time.Duration(ttlSeconds) * time.Second)
		if err := txn.SetEntry(entry); err != nil {
			return d.logger.LogAndError("failed to set alert status with TTL in DB: %v", "error", err)
		}
		return nil
	})
}

// GetAlertStatus 获取告警状态
func (d *AlertStatusManager) GetAlertStatus(host string) (AlertStatus, error) {
	var status AlertStatus
	key := d.generateKey(host)
	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {
			return json.Unmarshal(v, &status)
		})
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			// 记录不存在是正常情况，返回特定错误便于上层函数识别
			return AlertStatus{}, badger.ErrKeyNotFound
		}
		return AlertStatus{}, fmt.Errorf("failed to get alert status for host %s: %w", host, err)
	}
	return status, nil
}

// DeleteAlertStatus 删除告警状态
func (d *AlertStatusManager) DeleteAlertStatus(host string) error {
	key := d.generateKey(host)
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// MarkAsAlert 将主机状态标记为 ALERT
func (d *AlertStatusManager) MarkAsAlert(status AlertStatus) error {
	// 获取数据库中现有的状态
	existingStatus, err := d.GetAlertStatus(status.Host)
	if err != nil {
		// 如果是 Key 不存在的错误，直接插入新状态
		if err == badger.ErrKeyNotFound {
			d.logger.Log(fmt.Sprintf("Creating new alert status record for host: %s", status.Host), "debug")
			return d.SetAlertStatus(status, d.dbConfig.Expire)
		}
		// 其他错误直接返回
		return fmt.Errorf("failed to get alert status: %w", err)
	}

	// 如果数据库中状态已经是 ALERT，则无需更新
	if existingStatus.Status == StatusAlert {
		d.logger.Log(fmt.Sprintf("Host %s is already in ALERT state, skipping update", status.Host), "debug")
		return nil
	}

	// 如果数据库中状态是 RECOVERY，则更新为传入的完整状态
	d.logger.Log(fmt.Sprintf("Updating host %s from RECOVERY to ALERT", status.Host), "debug")
	return d.SetAlertStatus(status, d.dbConfig.Expire)
}

// GetAllUnsentStatuses 获取所有未发送的状态，根据传入的 Status 筛选
func (d *AlertStatusManager) GetAllUnsentStatuses(statusType StatusType) ([]*AlertStatus, error) {
	var statuses []*AlertStatus
	err := d.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()

		for iter.Rewind(); iter.Valid(); iter.Next() {
			item := iter.Item()
			var status AlertStatus
			err := item.Value(func(v []byte) error {
				return json.Unmarshal(v, &status)
			})
			if err != nil {
				return err
			}

			// 过滤条件：sent 为 false 且 status 为指定类型
			if !status.Sent && status.Status == statusType {
				statuses = append(statuses, &status) // 使用指针
			}
		}
		return nil
	})
	return statuses, err
}

// MarkAsRecovered 标记主机为已恢复状态
func (d *AlertStatusManager) MarkAsRecovered(status AlertStatus) error {
	existingStatus, err := d.GetAlertStatus(status.Host)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			// 如果没有之前的告警记录，则不需要恢复通知
			return nil
		}
		return fmt.Errorf("failed to get alert status: %w", err)
	}

	// 如果之前是 RECOVERY 状态，跳过数据库操作
	if existingStatus.Status == StatusRecovery {
		d.logger.Log(fmt.Sprintf("Host %s is already in RECOVERY state, skipping update", status.Host), "debug")
		return nil
	}

	// 如果之前是 ALERT 状态，更新为 RECOVERY 状态并重置 sent 为 false
	if existingStatus.Status == StatusAlert {
		d.logger.Log(fmt.Sprintf("Marking host %s as RECOVERY", status.Host), "debug")
		existingStatus.Status = StatusRecovery            // 更新为恢复状态
		existingStatus.Sent = false                       // 恢复通知未发送
		existingStatus.RecoveryTime = status.RecoveryTime // 设置恢复时间
		return d.SetAlertStatus(existingStatus, d.dbConfig.Expire)
	}

	// 如果状态是其他未知状态，记录警告日志并跳过
	d.logger.Log(fmt.Sprintf("Unhandled status for host %s: %s, skipping update", status.Host, existingStatus.Status), "warn")
	return nil
}

// UpdateSentStatus 更新主机的 Sent 状态
func (d *AlertStatusManager) UpdateSentStatus(host string, sent bool) error {
	existingStatus, err := d.GetAlertStatus(host)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			// 如果记录不存在，直接返回错误
			return fmt.Errorf("alert status for host %s not found", host)
		}
		return fmt.Errorf("failed to get alert status for host %s: %w", host, err)
	}

	// 更新 Sent 字段
	existingStatus.Sent = sent

	// 保存更新后的状态
	return d.SetAlertStatus(existingStatus, d.dbConfig.Expire)
}
