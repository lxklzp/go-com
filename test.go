package main

import (
	"fmt"
	queue "github.com/confluentinc/confluent-kafka-go/kafka"
	"go-com/config"
	"go-com/global"
	"time"
)

func main() {
	config.Load()
	global.InitLog("test")
	cfg := config.Kafka{
		Servers:          "192.168.2.70:9092",
		Username:         "",
		Password:         "",
		Topic:            "test_topic_num_",
		Group:            "",
		SecurityProtocol: "",
		SaslMechanisms:   "",
	}
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
		global.Log.Error(err)
	}

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *queue.Message:
				if ev.TopicPartition.Error != nil {
					global.Log.Errorf("[kafka] 消息投递失败: %v\n", ev.TopicPartition)
				} else {
					global.Log.Infof("[kafka] 消息投递成功: %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	data := []byte{1}
	// Produce messages to topic (asynchronously)
	for i := 0; i < 100000; i++ {
		topic := fmt.Sprintf("%s%d", cfg.Topic, i)
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
				global.Log.Errorf("[kafka] 消息投递失败: %v\n", err)
			}
			break
		}
	}

	// Wait for message deliveries before shutting down
	for p.Flush(10000) > 0 {
		global.Log.Info("[kafka] Still waiting to flush outstanding messages\n")
	}
	p.Close()
}
