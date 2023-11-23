package delay_queue

import (
	"context"
	"go-com/config"
	"go-com/global"
	"go-com/internal/model"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

var EtcdDQ etcdDQ

const DelayQueueFlag = "DQ"

type etcdDQ struct {
}

// DelayQueueProducer 将对应服务的数据序列号投递到延迟队列，秒
func (dq etcdDQ) DelayQueueProducer(serviceName string, seq string, expire int64) {
	ctx := context.TODO()
	key := config.C.App.Prefix + DelayQueueFlag + serviceName + seq
	lease := clientv3.NewLease(global.Etcd)
	defer lease.Close()
	leaseGrant, err := lease.Grant(context.Background(), expire)
	if err != nil {
		global.Log.Error(err)
	}
	if _, err = global.Etcd.Put(ctx, key, "", clientv3.WithLease(leaseGrant.ID)); err != nil {
		global.Log.Error(err)
	}
}

// OrderObToDelayQueue 在服务启动时调用 从数据库中取出因宕机未被消费掉的数据序列号，并重新投递到延迟队列
func (dq etcdDQ) OrderObToDelayQueue(serviceName string) {
	orderTableName := serviceName + "_order"
	var orders []struct {
		OrderNo string
	}
	deadline := time.Now().Format(global.DateTimeFormatter)

	page := 0
	global.GormPg.Table(orderTableName).Select("order_no").Where("expire_send_time<=?", deadline).Order("id desc").Limit(model.MaxPageRead).Find(&orders)
	for ; len(orders) > 0; global.GormPg.Table(orderTableName).Select("order_no").Where("expire_send_time<=?", deadline).Order("id desc").Limit(model.MaxPageRead).Offset(page * model.MaxPageRead).Find(&orders) {
		for _, order := range orders {
			dq.DelayQueueProducer(serviceName, order.OrderNo, 3)
		}
		page++
		orders = nil
	}
}

// DelayQueueConsumer 延迟队列消费者
func (dq etcdDQ) DelayQueueConsumer(serviceName string) {
	orderTableName := serviceName + "_order"
	watcher := clientv3.NewWatcher(global.Etcd)
	defer watcher.Close()

	keyPrefix := config.C.App.Prefix + DelayQueueFlag + serviceName
	for true {
		watchRespChan := watcher.Watch(context.TODO(), keyPrefix, clientv3.WithPrefix())
		for watchResp := range watchRespChan {
			for _, event := range watchResp.Events {
				switch event.Type {
				case mvccpb.DELETE:
					orderNo := strings.TrimPrefix(string(event.Kv.Key), keyPrefix)
					go func() {
						defer func() {
							if err := recover(); err != nil {
								global.Log.Error(global.ErrorStack(err))
							}
							<-global.DelayQueueConsumeWorkerNumCh
						}()

						var err error
						var orderOb map[string]interface{}
						if err = global.GormPg.Transaction(func(tx *gorm.DB) error {
							// 利用数据库锁处理并发消费问题
							if err = tx.Table(orderTableName).Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no=?", orderNo).Take(&orderOb).Error; err != nil {
								return err
							}
							return nil
						}); err != nil {
							global.Log.Error(err)
						} else {
							if len(orderOb) > 0 {
								// 业务逻辑
								switch serviceName {
								}
							}
						}
					}()
				}
			}
		}
	}
}
