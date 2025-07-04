package kafka

import (
	"context"
	queue "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/tool"
	"time"
)

type Config struct {
	config.Kafka
}

type Kafka struct {
	Consumer *queue.Consumer
	Producer *queue.Producer
	Admin    *queue.AdminClient
	cfgP     Config         // 生产者配置缓存
	cfgC     Config         // 消费者配置缓存
	cfgA     Config         // 管理员配置缓存
	L        *logrus.Logger // 消费的消息日志，根据config.Kafka.LogExpire判断是否写日志
}

// InitConsumer offset：earliest、latest
func (kafka *Kafka) InitConsumer(cfg Config, offset string) {
	kafka.cfgC = cfg
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
		logr.L.Fatal(err)
	}

	// 订阅主题
	err = kafka.Consumer.SubscribeTopics([]string{cfg.Topic}, nil)
	if err != nil {
		logr.L.Fatal(err)
	}
	logr.L.Infof("[kafka] 消费者连接到kafka并订阅主题%s，等待消息...", cfg.Topic)
	if cfg.LogExpire > 0 {
		kafka.L = logr.NewLog("kafka_"+cfg.Topic, false, cfg.LogExpire)
	}
}

func (kafka *Kafka) Consume(handler func(key []byte, msg []byte, timestamp *time.Time)) {
	event := kafka.Consumer.Poll(1000) // 阻塞1秒
	if event == nil {
		return
	}

	switch e := event.(type) {
	case *queue.Message:
		if kafka.cfgC.LogExpire > 0 {
			kafka.L.Infof("[%s] [%s] [%s] %s", time.Now().Format(config.DateTimeFormatter), e.TopicPartition, string(e.Key), string(e.Value))
		}
		// 处理消息
		key := e.Key
		value := e.Value
		timestamp := e.Timestamp
		defer func() {
			if err := recover(); err != nil {
				tool.ErrorStack(err)
			}

			// 根据auto.commit.interval.ms配置自动提交消费者offset
			_, err := kafka.Consumer.StoreMessage(e)
			if err != nil {
				logr.L.Errorf("[kafka] 消费者 StoreMessage错误 %+v", err)
			}
		}()
		handler(key, value, &timestamp)
	case queue.Error:
		logr.L.Errorf("[kafka] 消费者 错误 %+v", e)
	}
}

func (kafka *Kafka) CloseConsumer() {
	kafka.Consumer.Close()
}

func (kafka *Kafka) InitProducer(cfg Config) {
	var err error
	kafka.cfgP = cfg
	// 建立连接
	cfgMap := queue.ConfigMap{
		"bootstrap.servers": cfg.Servers,
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

	// 创建生产者
	kafka.Producer, err = queue.NewProducer(&cfgMap)
	if err != nil {
		logr.L.Fatal(err)
	}
}

func (kafka *Kafka) Produce(key []byte, data []byte, topic string) {
	if kafka.cfgP.Servers == "" {
		return
	}

	var err error
	// Delivery report handler for produced messages
	go func() {
		for e := range kafka.Producer.Events() {
			switch ev := e.(type) {
			case *queue.Message:
				if ev.TopicPartition.Error != nil {
					logr.L.Errorf("[kafka] 消息投递失败: %v\n", ev.TopicPartition)
				} else {
					logr.L.Debugf("[kafka] 消息投递成功: %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	// Produce messages to topic (asynchronously)
	if topic == "" {
		topic = kafka.cfgP.Topic
	}
	for {
		err = kafka.Producer.Produce(&queue.Message{
			TopicPartition: queue.TopicPartition{Topic: &topic, Partition: queue.PartitionAny},
			Key:            key,
			Value:          data,
		}, nil)
		if err != nil {
			if err.(queue.Error).Code() == queue.ErrQueueFull {
				// Producer queue is full, wait 1s for messages
				// to be delivered then try again.
				time.Sleep(time.Second)
				continue
			}
			logr.L.Errorf("[kafka] 消息投递失败: %v\n", err)
		}
		break
	}

	// Wait for message deliveries before shutting down
	for kafka.Producer.Flush(10000) > 0 {
		logr.L.Info("[kafka] Still waiting to flush outstanding messages\n")
	}
}

func (kafka *Kafka) CloseProducer() {
	kafka.Producer.Close()
}

func (kafka *Kafka) InitAdmin(cfg Config) error {
	var err error
	kafka.cfgA = cfg
	// 建立连接
	cfgMap := queue.ConfigMap{
		"bootstrap.servers": cfg.Servers,
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

	kafka.Admin, err = queue.NewAdminClient(&cfgMap)
	if err != nil {
		logr.L.Error(err)
		return err
	}
	return nil
}

func (kafka *Kafka) CloseAdmin() {
	kafka.Admin.Close()
}

// Offset 偏移量
type Offset struct {
	Partition   int `json:"partition"`
	OffsetMsg   int `json:"offset_msg"`
	OffsetGroup int `json:"offset_group"`
}

func (kafka *Kafka) AdminBacklog(partitionCount int) (map[int]Offset, error) {
	// 初始化数据
	ctx := context.TODO()
	offset := make(map[int]Offset)
	topicPartitionOffsets := make(map[queue.TopicPartition]queue.OffsetSpec)
	partitions := make([]queue.TopicPartition, 0, partitionCount)
	for i := 0; i < partitionCount; i++ {
		topicPartitionOffsets[queue.TopicPartition{
			Topic:     &kafka.cfgA.Topic,
			Partition: int32(i),
		}] = queue.LatestOffsetSpec

		partitions = append(partitions, queue.TopicPartition{
			Topic:     &kafka.cfgA.Topic,
			Partition: int32(i),
		})

		offset[i] = Offset{
			Partition:   i,
			OffsetMsg:   0,
			OffsetGroup: 0,
		}
	}

	// 获取消息偏移量
	resMsg, err := kafka.Admin.ListOffsets(ctx, topicPartitionOffsets, queue.SetAdminIsolationLevel(queue.IsolationLevelReadCommitted))
	if err != nil {
		logr.L.Error(err)
		return nil, err
	}

	// 获取消费者组偏移量
	gps := []queue.ConsumerGroupTopicPartitions{
		{
			Group:      kafka.cfgA.Group,
			Partitions: partitions,
		},
	}
	resGroup, err := kafka.Admin.ListConsumerGroupOffsets(ctx, gps, queue.SetAdminRequireStableOffsets(true))
	if err != nil {
		logr.L.Error(err)
		return nil, err
	}
	if len(resGroup.ConsumerGroupsTopicPartitions) == 0 {
		return nil, errors.New("没有查询到消费者组的偏移量")
	}

	// 生成偏移量数据
	for k, v := range resMsg.ResultInfos {
		o := offset[int(k.Partition)]
		o.OffsetMsg = int(v.Offset)
		offset[int(k.Partition)] = o
	}
	for _, v := range resGroup.ConsumerGroupsTopicPartitions[0].Partitions {
		o := offset[int(v.Partition)]
		o.OffsetGroup = int(v.Offset)
		offset[int(v.Partition)] = o
	}
	return offset, nil
}
