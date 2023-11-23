package global

import (
	queue "github.com/confluentinc/confluent-kafka-go/kafka"
	"go-com/config"
	"time"
)

type Kafka struct {
	Consumer *queue.Consumer
}

/*
kafka日志目录：/data/logs/kafka/kafka.out
*/

// InitConsumer offset：earliest、latest
func (kafka *Kafka) InitConsumer(cfg config.Kafka, offset string) {
	// 创建kafka消费者
	var err error
	cfgMap := queue.ConfigMap{
		"bootstrap.servers":             cfg.Servers,
		"group.id":                      cfg.Group,
		"auto.offset.reset":             offset, // 把位移重设到当前最新位移处，避免重复消费
		"enable.auto.offset.store":      false,  // 手动存储偏移量
		"session.timeout.ms":            6000,
		"heartbeat.interval.ms":         2000,
		"partition.assignment.strategy": "range",
	}
	if cfg.Username != "" {
		cfgMap["security.protocol"] = cfg.SecurityProtocol
		if cfgMap["security.protocol"] == "" {
			cfgMap["security.protocol"] = "SASL_PLAINTEXT"
		}
		cfgMap["sasl.mechanisms"] = cfg.SaslMechanisms
		if cfgMap["sasl.mechanisms"] == "" {
			cfgMap["sasl.mechanisms"] = "PLAIN"
		}
		cfgMap["sasl.username"] = cfg.Username
		cfgMap["sasl.password"] = cfg.Password
	}
	kafka.Consumer, err = queue.NewConsumer(&cfgMap)

	if err != nil {
		Log.Fatal(err)
	}

	// 订阅主题
	err = kafka.Consumer.SubscribeTopics([]string{cfg.Topic}, nil)
	if err != nil {
		Log.Fatal(err)
	}
	Log.Infof("[kafka] 消费者连接到kafka并订阅主题%s，等待消息...", cfg.Topic)
}

func (kafka *Kafka) Consume(handler func(msg []byte, timestamp *time.Time)) {
	event := kafka.Consumer.Poll(1000) // 阻塞1秒
	if event == nil {
		return
	}

	switch e := event.(type) {
	case *queue.Message:
		if config.C.App.KafkaToLog {
			Log.Infof("[kafka] consume %s: %s", e.TopicPartition, string(e.Value))
		}
		// 处理消息
		KafkaConsumeWorkerNumCh <- true
		go func() {
			e := e
			value := e.Value
			timestamp := e.Timestamp
			defer func() {
				if err := recover(); err != nil {
					Log.Error(err)
				}
				<-KafkaConsumeWorkerNumCh

				// 根据auto.commit.interval.ms配置自动提交消费者offset
				_, err := kafka.Consumer.StoreMessage(e)
				if err != nil {
					Log.Errorf("[kafka] 消费者 StoreMessage错误 %+v", err)
				}
			}()
			handler(value, &timestamp)
		}()
	case queue.Error:
		Log.Errorf("[kafka] 消费者 错误 %+v", e)
	}
}

func (kafka *Kafka) Produce(cfg config.Kafka, data []byte) {
	// 建立连接
	cfgMap := queue.ConfigMap{
		"bootstrap.servers": cfg.Servers,
	}
	if cfg.Username != "" {
		cfgMap["security.protocol"] = "SASL_PLAINTEXT"
		cfgMap["sasl.mechanisms"] = "SCRAM-SHA-512"
		cfgMap["sasl.username"] = cfg.Username
		cfgMap["sasl.password"] = cfg.Password
	}

	// 创建生产者
	p, err := queue.NewProducer(&cfgMap)
	if err != nil {
		Log.Error(err)
	}

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *queue.Message:
				if ev.TopicPartition.Error != nil {
					Log.Errorf("[kafka] 消息投递失败: %v\n", ev.TopicPartition)
				} else {
					Log.Infof("[kafka] 消息投递成功: %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	// Produce messages to topic (asynchronously)
	topic := cfg.Topic
	for {
		err = p.Produce(&queue.Message{
			TopicPartition: queue.TopicPartition{Topic: &topic, Partition: queue.PartitionAny},
			Value:          data,
		}, nil)
		if err != nil {
			if err.(queue.Error).Code() == queue.ErrQueueFull {
				// Producer queue is full, wait 1s for messages
				// to be delivered then try again.
				time.Sleep(time.Second)
				continue
			}
			Log.Errorf("[kafka] 消息投递失败: %v\n", err)
		}
		break
	}

	// Wait for message deliveries before shutting down
	for p.Flush(10000) > 0 {
		Log.Info("[kafka] Still waiting to flush outstanding messages\n")
	}
	p.Close()
}
