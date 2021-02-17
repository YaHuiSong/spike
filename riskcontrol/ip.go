package riskcontrol

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

//RequestIP  ip
type RequestIP struct {
	IP string
}

// RiskControl  redis key exist or not
func (req *RequestIP) RiskControl(c redis.Conn, ipKey string) bool {
	//判断key是否存在
	isKeyExit, err := redis.Bool(c.Do("EXISTS", ipKey))
	if err != nil {
		fmt.Printf("get error %v when get key from redis\n", err)
	}
	return isKeyExit
}

//GetKey redis key
func (req *RequestIP) GetKey(product string, dd time.Time) string {
	timeStr := dd.UTC().Format(timeLayout)
	return product + "_" + req.IP + "_" + timeStr
}
