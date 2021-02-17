package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	mrand "math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"example.com/spike-01/localspike"
	"example.com/spike-01/mysqlserver"
	"example.com/spike-01/rabbitmq"
	"example.com/spike-01/remotespike"
	"example.com/spike-01/riskcontrol"
	"example.com/spike-01/spikestatus"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
)

var (
	localSpike  *localspike.LocalSpike
	redisSpike  *remotespike.RemoteSpike
	redisPool   *redis.Pool
	mqServer    *rabbitmq.RabbitMQ
	mySQLServer *sql.DB
	done        chan int
)

const redisPort = "127.0.0.1:6379"

var spikeTime = time.Date(2021, 1, 23, 10, 0, 0, 0, time.Local)

func init() {
	localSpike = &localspike.LocalSpike{
		Stock: 150,
		Sales: 0,
	}
	redisSpike = redisSpike.GenereteKeysName(productID, spikeTime)
	redisPool = remotespike.NewPool(redisPort)
	mqServer = rabbitmq.NewRabbitMQ("spike", "", "spike")
	mySQLServer = mysqlserver.Instance("root", "alice", "127.0.0.1:3306", "spike")
	done = make(chan int, 1)
	done <- 1
}

func main() {
	// Disable Console Color, you don't need console color when writing the logs to file.
	gin.DisableConsoleColor()

	// Logging to a file.
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)

	router := gin.Default()
	router.POST("/spike", handleReq)
	router.GET("/status", handlerStatus)

	go func() {
		order()
	}()
	router.Run()
}

func handlerStatus(c *gin.Context) {
	redisConn := redisPool.Get()
	defer redisConn.Close()
	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
	}
	// productID := req["productID"].(string)
	// userID := req["userID"].(string)
	// userIP := req["userIP"].(string)

	msg := spikestatus.QueryStatus(spikeTime, redisSpike.GetRemain(redisConn))
	c.JSON(http.StatusOK, gin.H{
		"status": msg,
	})
}

func riskFilters(productID, userID, userIP string, redisConn redis.Conn) bool {
	riskControls := []riskcontrol.RiskControl{&riskcontrol.RequestUser{UserID: userID}, &riskcontrol.RequestIP{IP: userIP}}
	for _, riskControl := range riskControls {
		if !validate(riskControl, redisConn, productID, spikeTime) {
			return false
		}
	}
	return true
}

func handleReq(c *gin.Context) {
	redisConn := redisPool.Get()
	defer redisConn.Close()
	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		// return 需不需要？？？？
	}
	productID := req["productID"].(string)
	userID := req["userID"].(string)
	userIP := req["userIP"].(string)
	logMsg := ""
	<-done
	fmt.Println(fmt.Sprintf("active count: %d; idle count: %d; wait count: %d", redisPool.Stats().ActiveCount, redisPool.Stats().IdleCount, redisPool.Stats().WaitCount))
	if !riskFilters(productID, userID, userIP, redisConn) {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "不能重复抢",
		})
	}
	if localSpike.DeductStock() && redisSpike.DeductionSales(redisConn) {
		// send order to mq
		// Create a new Node with a Node number of 1
		node, err := snowflake.NewNode(1)
		if err != nil {
			logMsg = logMsg + "generate global unique id failed for new order"
		}
		// Generate a snowflake ID.
		id := node.Generate()
		mrand.Seed(time.Now().UnixNano())
		svr := mysqlserver.NewOrder("coupon_spike", id.String(), productID, userID)
		msg, err := json.Marshal(svr)
		if err != nil {
			logMsg = logMsg + "创建订单失败！"
		}
		flag := mqServer.PublishSimple(msg)
		if flag {
			c.JSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "抢券成功",
			})
			logMsg = logMsg + "result:1,localSales:" + strconv.FormatInt(localSpike.Sales, 10)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": "抢券失败",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": "已售罄",
		})
		logMsg = logMsg + "result:0,localSales:" + strconv.FormatInt(localSpike.Sales, 10)
	}
	done <- 1
	writeLog(logMsg, "./stat.log")
}

func validate(risk riskcontrol.RiskControl, c redis.Conn, productID string, dd time.Time) bool {
	key := risk.GetKey(productID, dd)
	return risk.RiskControl(c, key)
}

func writeLog(msg string, logPath string) {
	fd, _ := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	defer fd.Close()
	content := strings.Join([]string{msg, "\r\n"}, "")
	buf := []byte(content)
	fd.Write(buf)
}

func handlerMQMsg(msg []byte) bool {
	// uid是否合法
	// 库存是否充足

	stockObj := mysqlserver.NewStock("productions", productID)
	if stockObj.DeductionStore(mySQLServer) {
		order := mysqlserver.NewOrder("", "", "", "")
		err := json.Unmarshal(msg, order)
		if err != nil {
			writeLog("mesage to json failed", "./stat.log")
		}
		if order.AddOrder(mySQLServer) {
			return true
		}
		writeLog("add order failed with "+string(msg), "./stat.log")
		return false
	}
	return false
}

func order() {
	mqServer.ConsumeSimple(handlerMQMsg)
}
