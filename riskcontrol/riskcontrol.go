package riskcontrol

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

const timeLayout = "20060102_15UTC"

//RiskControl intterface
type RiskControl interface {
	GetKey(string, time.Time) string
	RiskControl(redis.Conn, string) bool
}

//RiskControlConfig config
type Config struct {
	RiskList []string `json:"RiskList"`
}

//RiskFilter filter risk
// func RiskFilter(src string, dd time.Time) {
// 	data, err := ioutil.ReadFile(src)
// 	if err != nil {
// 		panic(err)
// 	}
// 	conf := &Config{}
// 	err = json.Unmarshal(data, &conf)
// 	if err != nil {
// 		panic(err)
// 	}
// 	var r RiskControl
// 	for _, v := range conf.RiskList {
// 		r = v{}
// 	}
// }
