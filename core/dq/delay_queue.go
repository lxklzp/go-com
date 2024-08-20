package dq

import (
	"container/heap"
	"encoding/json"
	"go-com/config"
	"go-com/core/filer"
	"go-com/core/logr"
	"go-com/core/tool"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// 延迟队列，支持生产、消费、定时持久化

type Config struct {
	config.Dq
}

// Message 延迟队列里面的一条消息
type Message struct {
	Timestamp int64       // unix时间戳，到这个时间就出队
	Topic     string      // 主题，区分不同类型的消息
	No        string      // 消息唯一编号
	Data      interface{} // 消息数据
}

// Queue 队列
type Queue struct {
	list        []Message
	lock        sync.Mutex
	filename    string // 持久化文件名
	workerNumCh chan bool
	runningNo   sync.Map // 执行中的消息的no列表
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
		list:        []Message{},
		lock:        sync.Mutex{},
		filename:    filename,
		workerNumCh: make(chan bool, config.C.Dq.MaxWorkerNum),
	}
	heap.Init(q)
	return q
}

func (q *Queue) Produce(msg Message) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if config.C.Dq.CheckNoExist {
		for _, m := range q.list {
			if m.No == msg.No {
				return
			}
		}
	}

	heap.Push(q, msg)
}

func (q *Queue) Consume(handler func(msg Message)) {
	for {
		q.lock.Lock()
		if q.Len() == 0 {
			q.lock.Unlock()
			break
		}

		m := heap.Pop(q).(Message)

		// 判断执行中的消息的no是否存在
		if _, ok := q.runningNo.Load(m.No); ok && config.C.Dq.CheckNoRunningExist {
			q.lock.Unlock()
			continue
		}
		q.runningNo.Store(m.No, true)
		q.lock.Unlock()

		if m.Timestamp <= time.Now().Unix() {
			if config.C.Dq.MaxWorkerNum == 0 {
				// 0表示串行
				handler(m)
				q.runningNo.Delete(m.No)
			} else {
				// 并行
				q.workerNumCh <- true
				go func() {
					defer func() {
						if err := recover(); err != nil {
							logr.L.Error(tool.ErrorStack(err))
						}
						q.runningNo.Delete(m.No)
						<-q.workerNumCh
					}()
					handler(m)
				}()
			}
		} else {
			q.lock.Lock()
			heap.Push(q, m)
			q.runningNo.Delete(m.No)
			q.lock.Unlock()
			break
		}
	}
}

func (q *Queue) Run(handler func(msg Message)) {
	var err error
	// 载入持久化文件
	if config.C.Dq.PersistPeriod > 0 && filer.Exist(q.filename) {
		listJson, _ := os.ReadFile(q.filename)
		if err = json.Unmarshal(listJson, &q.list); err != nil {
			logr.L.Fatal(err)
		}
	}

	go q.loopConsume(handler)
	if config.C.Dq.PersistPeriod > 0 {
		go q.loopPersist()
	}
}

func (q *Queue) Persist() {
	var err error
	q.lock.Lock()
	listJson, _ := json.Marshal(q.list)
	q.lock.Unlock()
	if err = os.WriteFile(q.filename, listJson, 0755); err != nil {
		logr.L.Error(err)
	}
}

// 轮询消费
func (q *Queue) loopConsume(handler func(msg Message)) {
	ticker := time.NewTicker(time.Second * time.Duration(config.C.Dq.ConsumePeriod))
	defer func() {
		if err := recover(); err != nil {
			logr.L.Error(tool.ErrorStack(err))
		}
		ticker.Stop()
		go q.loopConsume(handler)
	}()

	for range ticker.C {
		q.Consume(handler)
	}
}

// 轮询持久化
func (q *Queue) loopPersist() {
	ticker := time.NewTicker(time.Second * time.Duration(config.C.Dq.PersistPeriod))
	defer func() {
		if err := recover(); err != nil {
			logr.L.Error(err)
		}
		ticker.Stop()
		go q.loopPersist()
	}()

	for range ticker.C {
		q.Persist()
	}
}

// CycleTime 轮询时间
type CycleTime struct {
	BeginTime string        // 开始时间
	Type      int           // 1 循环 2 定时
	Period    time.Duration // 循环时间间隔，单位分钟
	TimePoint []string      // 定时时间点数组，12:34:56格式
}

// NextTime 周期性任务时，获取下一次执行时间，period单位：分钟
func (q *Queue) NextTime(cycleTime CycleTime) time.Time {
	now := time.Now()
	var nextTime time.Time
	switch cycleTime.Type {
	case 1:
		nextTime, _ = time.ParseInLocation(config.DateTimeFormatter, cycleTime.BeginTime, time.Local)
		// 下一次时间要在当前时间之后
		offset := now.Sub(nextTime).Minutes() / float64(cycleTime.Period)
		if offset > 0 {
			nextTime = nextTime.Add(time.Minute * cycleTime.Period * time.Duration(int(offset)+1))
		}
		return nextTime
	case 2:
		sort.Strings(cycleTime.TimePoint)
		datePointBegin := cycleTime.BeginTime[:10]
		datePointNow := now.Format(config.DateFormatter)
		// 开始时间在今天以后
		if strings.Compare(datePointBegin, datePointNow) > 0 {
			nextTime, _ = time.ParseInLocation(config.DateTimeFormatter, datePointBegin+" "+cycleTime.TimePoint[0], time.Local)
			return nextTime
		}
		// 开始时间在今天或者今天以前
		timePoint := ""
		for _, t := range cycleTime.TimePoint {
			// 找到今天的下一个时间点
			if strings.Compare(t, now.Format(config.TimeFormatter)) > 0 {
				timePoint = t
				break
			}
		}
		if timePoint == "" {
			// 今天没有时间点了，改成明天
			nextTime, _ = time.ParseInLocation(config.DateTimeFormatter, datePointNow+" "+cycleTime.TimePoint[0], time.Local)
			nextTime = nextTime.Add(time.Hour * 24)
		} else {
			// 今天有时间点
			nextTime, _ = time.ParseInLocation(config.DateTimeFormatter, datePointNow+" "+timePoint, time.Local)
		}
		return nextTime
	case 3:
		return now
	default:
		return time.Time{}
	}
}
