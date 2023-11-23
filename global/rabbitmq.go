package global

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"go-com/config"
	"time"
)

type Rabbitmq struct {
	Conn *amqp.Connection
}

// 建立连接，并发安全
func (mq *Rabbitmq) Connection() {
	cfg := config.C.Rabbitmq
	var err error
	mq.Conn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/", cfg.User, cfg.Password, cfg.Addr))
	if err != nil {
		Log.Panic(err)
	}
}

// 通道不支持并发安全
func (mq *Rabbitmq) Chan() *amqp.Channel {
	ch, err := mq.Conn.Channel()
	if err != nil {
		Log.Panic(err)
	}
	return ch
}

// 初始化延迟队列
func (mq *Rabbitmq) InitDelayQueue(ch *amqp.Channel, queueName string, exName string, routeKey string) {
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		Log.Panic(err)
	}

	err = ch.ExchangeDeclare(
		exName,              // name
		"x-delayed-message", // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		amqp.Table{
			"x-delayed-type": "direct",
		}, // arguments
	)
	if err != nil {
		Log.Panic(err)
	}

	err = ch.QueueBind(
		q.Name,   // queue name
		routeKey, // routing key
		exName,   // exchange
		false,
		nil)
	if err != nil {
		Log.Panic(err)
	}

	Log.Infof("[rabbitmq] %s 延迟队列初始化成功", queueName)
}

// 消费队列消息
func (mq *Rabbitmq) Consume(ch *amqp.Channel, queueName string) <-chan amqp.Delivery {
	err := ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		Log.Panic(err)
	}

	msgCh, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto ack
		false,     // exclusive
		false,     // no local
		false,     // no wait
		nil,       // args
	)
	if err != nil {
		Log.Panic(err)
	}

	Log.Infof("[rabbitmq] %s 消费者初始化成功，等待消息...", queueName)
	return msgCh
}

// 投递消息到延迟队列
func (mq *Rabbitmq) ProduceDelayQueue(ch *amqp.Channel, exName string, routeKey string, msg []byte, delay int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ch.PublishWithContext(ctx,
		exName,   // exchange
		routeKey, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
			Headers: map[string]interface{}{
				"x-delay": 1000 * delay, // 消息从交换机过期时间,毫秒（x-dead-message插件提供）
			},
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		Log.Panic(err)
	} else {
		fmt.Printf("[rabbitmq] 成功发送%s消息到 %s:%s\n", msg, exName, routeKey)
	}
}

// 初始化优先级队列
func (mq *Rabbitmq) InitPriorityQueue(ch *amqp.Channel, queueName string, exName string, routeKey string) {
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		amqp.Table{
			"x-max-priority": 10,
			"x-queue-mode":   "lazy",
			"x-max-length":   10000,
		}, // arguments
	)
	if err != nil {
		Log.Panic(err)
	}

	err = ch.ExchangeDeclare(
		exName,   // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		Log.Panic(err)
	}

	err = ch.QueueBind(
		q.Name,   // queue name
		routeKey, // routing key
		exName,   // exchange
		false,
		nil)

	if err != nil {
		Log.Panic(err)
	}
	Log.Infof("[rabbitmq] %s 优先级队列初始化成功", queueName)
}

// 投递消息到优先级队列
func (mq *Rabbitmq) ProducePriorityQueue(ch *amqp.Channel, exName string, routeKey string, msg []byte, priority uint8) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ch.PublishWithContext(ctx,
		exName,   // exchange
		routeKey, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Priority:     priority + 1, // 项目是0-31，rabbitmq是1-32
			Body:         msg,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		Log.Error(err)
		return false
	} else {
		fmt.Printf("[rabbitmq] 成功发送%s消息到 %s:%s\n", msg, exName, routeKey)
		return true
	}
}

// 初始化队列
func (mq *Rabbitmq) InitDirectQueue(ch *amqp.Channel, queueName string, exName string, routeKey string) {
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		Log.Panic(err)
	}

	err = ch.ExchangeDeclare(
		exName,   // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		Log.Panic(err)
	}

	err = ch.QueueBind(
		q.Name,   // queue name
		routeKey, // routing key
		exName,   // exchange
		false,
		nil)
	if err != nil {
		Log.Panic(err)
	}

	Log.Infof("[rabbitmq] %s 延迟队列初始化成功", queueName)
}

// 投递消息到队列
func (mq *Rabbitmq) ProduceDirectQueue(ch *amqp.Channel, exName string, routeKey string, msg []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ch.PublishWithContext(ctx,
		exName,   // exchange
		routeKey, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         msg,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		Log.Panic(err)
	} else {
		fmt.Printf("[rabbitmq] 成功发送%s消息到 %s:%s\n", msg, exName, routeKey)
	}
}
