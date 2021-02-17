package mysqlserver

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

//MySQLServer instance
type MySQLServer struct {
	DriverName string
	User       string
	Pwd        string
	ServerIP   string
	DBName     string
}

// CouponOrder :
type CouponOrder struct {
	TableName string
	OrderID   string
	UserID    string
	ProductID string
	Status    string
}

// InStock in stock
type InStock struct {
	TableName string
	ProductID string
}

//NewOrder []byte new order
func NewOrder(tableName, orderID, productionID, userID string) *CouponOrder {
	return &CouponOrder{
		TableName: tableName,
		OrderID:   orderID,
		UserID:    userID,
		ProductID: productionID,
	}
}

//NewOrder []byte new order
func NewStock(tableName, productionID string) *InStock {
	return &InStock{
		TableName: tableName,
		ProductID: productionID,
	}
}

//Instance :
func Instance(user, pwd, server, dbName string) *sql.DB {
	m := &MySQLServer{
		DriverName: "mysql",
		User:       user,
		Pwd:        pwd,
		ServerIP:   server,
		DBName:     dbName,
	}
	dbURL := m.User + ":" + m.Pwd + "@tcp(" + m.ServerIP + ")/" + m.DBName
	var err error
	db, err = sql.Open(m.DriverName, dbURL)
	log.Printf("%s", "连接MySQL中。。。")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(300)

	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("mysql 连接成功")
	return db
}

//CloseConn Close Connection
func CloseConn() {
	db.Close()
}

//AddOrder :
func (order *CouponOrder) AddOrder(db *sql.DB) bool {
	// sqlStr := "insert into ? (`product_id`,`user_id`,`status`) values (? , ? , ?)"
	sqlStr := fmt.Sprintf(
		"insert into %s(order_id,product_id,user_id,status) values ('%s','%s', '%s', '%s')",
		order.TableName, order.OrderID, order.ProductID, order.UserID, "paying")
	log.Printf("正在执行sql 命令 %s", sqlStr)
	insertResult, err := db.Exec(sqlStr)
	if err != nil {
		log.Fatalln(sqlStr, err)
		return false
	}
	affected, _ := insertResult.RowsAffected()
	if affected == 1 {
		log.Printf("sql 命令 %s 执行成功", sqlStr)
		return true
	}
	log.Printf("sql 命令 %s 执行失败", sqlStr)
	return false
}

//DeductionStore reduce the num in stock
func (stock *InStock) DeductionStore(db *sql.DB) bool {
	// sqlStr := "update productions set `total_num = IF(`total_num`< 1, 0, `total_num`- 1) WHERE `product_id` = ?"
	sqlStr := fmt.Sprintf(
		"update %s set remain_num=remain_num - 1 where product_id=%s and remain_num > 0",
		stock.TableName,
		stock.ProductID,
	)
	log.Printf("正在执行sql 命令 %s", sqlStr)
	insertResult, err := db.Exec(sqlStr)
	if err != nil {
		log.Fatalln(err)
		return false
	}
	affected, _ := insertResult.RowsAffected()
	if affected == 1 {
		log.Printf("sql 命令 %s 执行成功", sqlStr)
		return true
	}
	log.Printf("sql 命令 %s 执行失败", sqlStr)
	return false
}

// log.Printf("get id ===> %d, affected =====> %d\n", id, affected)

// rows, err := db.Query("select id, name, email from chaos.mytb")

// if err != nil {
// 	log.Fatalln(err)
// }

// defer rows.Close()

// for rows.Next() {
// 	mytb := Mytb{}

// 	err = rows.Scan(&mytb.Id, &mytb.Name, &mytb.Email)
// 	if err != nil {
// 		log.Printf(">>>>>>>>>>> db %v\n", db)
// 		log.Fatalln(err)
// 	}
// 	log.Printf("found row containing ..%v %v %v \n", mytb.Id, mytb.Name, mytb.Email)

// }
