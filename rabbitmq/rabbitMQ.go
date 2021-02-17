package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// MQURL 格式 amqp://账号：密码@rabbitmq服务器地址：端口号/vhost
const MQURL = "amqp://guest:guest@127.0.0.1:5672"

//RabbitMQ :
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	// 队列名称
	QueueName string
	// 交换机
	Exchange string
	// Key
	Key string
	// 连接信息
	Mqurl string
}

// NewRabbitMQ 创建结构体实例
func NewRabbitMQ(queueName, exchange, key string) *RabbitMQ {
	rabbitmq := &RabbitMQ{
		QueueName: queueName,
		Exchange:  exchange,
		Key:       key,
		Mqurl:     MQURL,
	}
	var err error
	// 创建rabbitmq连接
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "创建连接错误！")

	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "获取channel失败！")

	return rabbitmq
}

// Destory 断开channel和connection
func (r *RabbitMQ) Destory() {
	_ = r.channel.Close()
	_ = r.conn.Close()
}

// failOnErr 错误处理函数
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
		panic(fmt.Sprintf("%s:%s", message, err))
	}
}

// NewRabbitMQSimple
// 简单模式Step 1.创建简单模式下的RabbitMq实例
func NewRabbitMQSimple(queueName string) *RabbitMQ {
	return NewRabbitMQ(queueName, "", "")
}

// PublishSimple 简单模式Step 2:简单模式下生产代码
func (r *RabbitMQ) PublishSimple(message []byte) bool {
	// 1. 申请队列，如果队列不存在会自动创建，如何存在则跳过创建
	// 保证队列存在，消息能发送到队列中
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		// 是否持久化
		true,
		// 是否为自动删除
		false,
		// 是否具有排他性
		false,
		// 是否阻塞
		false,
		// 额外属性
		nil,
	)
	if err != nil {
		r.failOnErr(err, "声明 queue失败！")
		return false
	}

	// 2.发送消息到队列中
	err = r.channel.Publish(
		r.Exchange,
		r.QueueName,
		// 如果为true, 会根据exchange类型和routkey规则，如果无法找到符合条件的队列那么会把发送的消息返回给发送者
		false,
		// 如果为true, 当exchange发送消息到队列后发现队列上没有绑定消费者，则会把消息发还给发送者
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         message,
		})
	if err != nil {
		r.failOnErr(err, "Failed to publish a message")
		return false
	} else {
		log.Printf(" [x] Sent %s", message)
		return true
	}
}

// ConsumeSimple 使用 goroutine 消费消息
func (r *RabbitMQ) ConsumeSimple(handler func([]byte) bool) {
	// 1. 申请队列，如果队列不存在会自动创建，如何存在则跳过创建
	// 保证队列存在，消息能发送到队列中
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		// 是否持久化
		true,
		// 是否为自动删除
		false,
		// 是否具有排他性
		false,
		// 是否阻塞
		false,
		// 额外属性
		nil,
	)
	if err != nil {
		r.failOnErr(err, "声明 queue失败!")
	}

	err = r.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		r.failOnErr(err, "设置QoS 失败!")
	}
	// 接收消息
	msgs, err := r.channel.Consume(
		r.QueueName,
		// 用来区分多个消费者
		"",
		// 是否自动应答 ？？？
		false,
		// 是否具有排他性
		false,
		// 如果设置为true，表示不能将同一个connection中发送的消息传递给这个connection中的消费者
		false,
		// 队列消费是否阻塞
		false,
		nil,
	)

	if err != nil {
		r.failOnErr(err, "注册消费者失败！")
	}

	forever := make(chan bool)
	// 启用协和处理消息
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			// 插入数据库
			suc := handler(d.Body)
			if suc {
				log.Printf("订单生成成功！")
			} else {
				log.Printf("订单生成失败！")
			}
			d.Ack(false)
		}
	}()
	log.Printf("[*] Waiting for message, To exit press CTRL+C")
	<-forever
}
