package riskcontrol

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

//UserRequest 请求用户的信息
type RequestUser struct {
	UserID string
}

//RiskControl  usr key exist in redis
func (req *RequestUser) RiskControl(c redis.Conn, userKey string) bool {
	//判断key是否存在
	isKeyExit, err := redis.Bool(c.Do("EXISTS", userKey))
	if err != nil {
		fmt.Printf("get error %v when get key from redis\n", err)
	}
	return isKeyExit
}

//GetKey redis key
func (req *RequestUser) GetKey(product string, dd time.Time) string {
	timeStr := dd.UTC().Format(timeLayout)
	return product + "_" + req.UserID + "_" + timeStr
}
