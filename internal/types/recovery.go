package types

import "time"

// RecoveryInfo 保存恢复通知所需信息
type RecoveryInfo struct {
	FailTime     time.Time // 故障发生时间
	RecoveryTime time.Time // 恢复时间
}
