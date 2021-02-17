package spikestatus

import "time"

// QueryStatus : 根据时间和库存返回目前能否抢购的状态
func QueryStatus(dd time.Time, total int) bool {
	if time.Now().Before(dd) || total == 0 {
		return false
	}
	return true
}
