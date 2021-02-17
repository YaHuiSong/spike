package remotespike

import (
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
)

//RemoteSpike the prudction keys
type RemoteSpike struct {
	ProductionTotalKey string
	ProductionSalesKey string
}

const timeLayout = "20060102_15UTC"

//GenereteKeysName : 根据production的名字生成key值
func (spike *RemoteSpike) GenereteKeysName(product string, dd time.Time) *RemoteSpike {
	timeStr := dd.UTC().Format(timeLayout)
	return &RemoteSpike{
		ProductionTotalKey: product + "_total_" + timeStr,
		ProductionSalesKey: product + "_sales_" + timeStr,
	}
}

//NewPool : 初始化redis连接池, redis port
func NewPool(port string) *redis.Pool {
	return &redis.Pool{
		MaxActive:   100,
		MaxIdle:     10,
		Wait:        true,
		IdleTimeout: 10 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", port)
			if err != nil {
				log.Printf("初始化redis线程池失败")
				panic(err.Error())
			}
			log.Printf("初始化redis线程池成功")
			return c, err
		},
	}
}

const luaScript = `
		local product_total_key = KEYS[1]
		local product_sales_key = KEYS[2]
		local user_order_status = KEYS[3]
		local product_user_key = KEYS[4]
		local productr_ip_key = KEYS[5]
		local user_exist = redis.call('get',product_user_key)
		local ip_exist = redis.call('get',productr_ip_key)
		if( not user_exist and not ip_exist) then
			return 0
		end
		local product_total_num = tonumber(redis.call('get',product_total_key))
		local product_sales_num = tonumber(redis.call('get',product_sales_key))

		if(product_sales_num < product_total_num) then
			return redis.call('INCR', product_sales_key) 
		end
		return 0
`

// local product_user_risk_list = redis.call('HMGET',user_order_status, user_id,user_ip)
// if(product_user_risk_list[0] || product_user_risk_list[1]) then
// 	return 0
// end
//DeductionSales : redis  库存 - 1
func (spike *RemoteSpike) DeductionSales(conn redis.Conn) bool {
	lua := redis.NewScript(5, luaScript)
	log.Printf("redis 更新已售卖的数量 lua 脚本 %s 执行中。。。", luaScript)
	re, err := redis.Int(lua.Do(conn, spike.ProductionTotalKey, spike.ProductionSalesKey))
	if err != nil {
		log.Printf("redis 更新已售卖的数量 lua 脚本 %s 出错 %v", luaScript, err)
		fmt.Println(err)
		return false
	}
	return re != 0
}

//TODO
//AddOrderUser : 添加或者更新状态
func (spike *RemoteSpike) AddOrderUser(conn redis.Conn, msg string) bool {
	lua := redis.NewScript(1, luaScript)
	re, err := redis.Int(lua.Do(conn, spike.ProductionTotalKey, spike.ProductionSalesKey, spike.BoughtUsersKey))
	if err != nil {
		return false
	}
	return re != 0
}

// GetRemain get the num in stock
func (spike *RemoteSpike) GetRemain(conn redis.Conn) int {
	total, err := redis.Int(conn.Do("Get", spike.ProductionTotalKey))
	if err != nil {
		fmt.Println("redis.read err=", err)
	}
	sale, err := redis.Int(conn.Do("Get", spike.ProductionSalesKey))
	if err != nil {
		fmt.Println("redis.read err=", err)
	}
	return total - sale
}
