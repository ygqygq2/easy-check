package types

import "time"

// EventTime 保存恢复通知所需信息
type EventTime struct {
	FailTime     time.Time // 故障发生时间
	RecoveryTime time.Time // 恢复时间
}
