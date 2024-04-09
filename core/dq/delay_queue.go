package dq

import (
	"container/heap"
	"encoding/json"
	"go-com/config"
	"go-com/core/filer"
	"go-com/core/logr"
	"log"
	"os"
	"sync"
	"time"
)

// Message 延迟队列里面的一条消息
type Message struct {
	Timestamp int64       // unix时间戳，到这个时间就出队
	Topic     string      // 主题，区分不同类型的消息
	No        int64       // 消息唯一编号
	Data      interface{} // 消息数据
}

// Queue 队列
type Queue struct {
	list     []Message
	lock     sync.Mutex
	filename string // 持久化文件名
}

/*---------- container/heap的实现 ----------*/

func (q *Queue) Len() int           { return len(q.list) }
func (q *Queue) Less(i, j int) bool { return q.list[i].Timestamp < q.list[j].Timestamp }
func (q *Queue) Swap(i, j int)      { q.list[i], q.list[j] = q.list[j], q.list[i] }
func (q *Queue) Push(x any) {
	(*q).list = append((*q).list, x.(Message))
}
func (q *Queue) Pop() any {
	old := (*q).list
	n := len(old)
	x := old[n-1]
	(*q).list = old[0 : n-1]
	return x
}

/*---------- 方法 ----------*/

func NewQueue() *Queue {
	// 持久化文件
	path := config.RuntimePath + "/delay_queue"
	if err := os.MkdirAll(path, 0777); err != nil {
		log.Fatal(err)
	}
	filename := path + "/list.json"

	q := &Queue{
		list:     []Message{},
		lock:     sync.Mutex{},
		filename: filename,
	}
	heap.Init(q)
	return q
}

func (q *Queue) Produce(m Message) {
	q.lock.Lock()
	defer q.lock.Unlock()

	heap.Push(q, m)
}

func (q *Queue) consume(handler func(m Message)) {
	for {
		q.lock.Lock()
		if q.Len() == 0 {
			q.lock.Unlock()
			break
		}
		m := heap.Pop(q).(Message)
		q.lock.Unlock()
		if m.Timestamp <= time.Now().Unix() {
			handler(m)
		} else {
			q.lock.Lock()
			heap.Push(q, m)
			q.lock.Unlock()
			break
		}
	}
}

func (q *Queue) Run(handler func(m Message)) {
	var err error
	// 载入持久化文件
	if filer.Exist(q.filename) {
		listJson, _ := os.ReadFile(q.filename)
		if err = json.Unmarshal(listJson, &q.list); err != nil {
			logr.L.Fatal(err)
		}
	}

	go q.loopConsume(handler)
	if config.C.App.DelayQueuePersistPeriod > 0 {
		go q.loopPersist()
	}
}

// 轮询消费
func (q *Queue) loopConsume(handler func(m Message)) {
	ticker := time.NewTicker(time.Second * time.Duration(config.C.App.DelayQueueConsumePeriod))
	defer func() {
		if err := recover(); err != nil {
			logr.L.Error(err)
		}
		ticker.Stop()
		go q.loopConsume(handler)
	}()

	for range ticker.C {
		q.consume(handler)
	}
}

// 轮询持久化
func (q *Queue) loopPersist() {
	ticker := time.NewTicker(time.Second * time.Duration(config.C.App.DelayQueuePersistPeriod))
	defer func() {
		if err := recover(); err != nil {
			logr.L.Error(err)
		}
		ticker.Stop()
		go q.loopPersist()
	}()

	var err error
	var listJson []byte
	for range ticker.C {
		q.lock.Lock()
		listJson, _ = json.Marshal(q.list)
		q.lock.Unlock()
		if err = os.WriteFile(q.filename, listJson, 0755); err != nil {
			logr.L.Error(err)
		}
	}
}
