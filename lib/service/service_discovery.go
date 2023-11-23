package service

import (
	"context"
	"go-com/config"
	"go-com/global"
	"go-com/lib/hash"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strconv"
	"strings"
	"sync"
)

var SD serviceDiscovery

const SDKeyPrefix = "sd-go-com/"
const SDApiPrefix = "api"
const SDMergePrefix = "merge"

const SDDefaultPrevAddr = "unknown_prev_addr"

// 服务发现 etcd的key结构：统一前缀/服务名称/服务器编号
type serviceDiscovery struct {
	cache sync.Map // map[string]SDService 服务名称->服务->服务器编号->服务器
}

// SDServer 服务器
type SDServer struct {
	serviceName string // 服务名称
	serverId    string // 服务器编号
	ServerAddr  string // 服务器地址
}

// SDService 服务
type SDService struct {
	List           map[string]SDServer  // 服务器列表 服务器编号->服务器
	index          int64                // 服务发现次数，单调递增
	consistentHash *hash.ConsistentHash // 一致性哈希，存储服务器编号
}

func (sd *serviceDiscovery) Registry(serviceName string, serverId string, ServerAddr string) {
	// 创建和声明一个租约，并且设置ttl（60秒）
	ctx := context.TODO()
	lease := clientv3.NewLease(global.Etcd)
	leaseGrant, err := lease.Grant(context.Background(), 60)
	if err != nil {
		global.Log.Fatal(err)
	}
	// 通过put记录SDServer，实现服务注册
	server := SDServer{
		serviceName: serviceName,
		serverId:    serverId,
		ServerAddr:  ServerAddr,
	}
	key := SDKeyPrefix + server.serviceName + "/" + server.serverId
	if _, err = global.Etcd.Put(ctx, key, server.ServerAddr, clientv3.WithLease(leaseGrant.ID)); err != nil {
		global.Log.Fatal(err)
	}

	defer func() {
		if err := recover(); err != nil {
			global.Log.Error(err)
		}
		lease.Close()
		sd.Registry(serviceName, serverId, ServerAddr)
	}()

	// 租约保活
	keepRespChan, err := lease.KeepAlive(ctx, leaseGrant.ID)
	if err != nil {
		global.Log.Fatal(err)
	}
	global.Log.Infof("etcd服务注册上线：%v", server)
	for range keepRespChan {
	}
	global.Log.Errorf("etcd服务注册掉线：%v", server)
}

// Watch 维护服务、服务器信息
func (sd *serviceDiscovery) Watch(serviceNamePrefix string) {
	ctx := context.TODO()

	// 通过get初始化列表
	getResp, err := global.Etcd.Get(ctx, SDKeyPrefix+serviceNamePrefix, clientv3.WithPrefix())
	if err != nil {
		global.Log.Fatal(err)
	}
	cache := make(map[string]*SDService)
	for _, v := range getResp.Kvs {
		key := strings.Split(string(v.Key), "/")
		serviceName := key[1]
		serverId := key[2]
		serverAddr := string(v.Value)
		server := SDServer{serviceName: serviceName, serverId: serverId, ServerAddr: serverAddr}
		if _, ok := cache[serviceName]; !ok {
			cache[serviceName] = &SDService{
				List:           make(map[string]SDServer),
				index:          0,
				consistentHash: hash.NewConsistentHash(),
			}
		}
		cache[serviceName].List[serverId] = server
		cache[serviceName].consistentHash.Add(serverId)
		global.Log.Infof("上线服务器 %s/%s:%s\n", serviceName, serverId, serverAddr)
	}
	for k, v := range cache {
		sd.cache.Store(k, v)
	}

	// 监听前缀key的值变化
	watcher := clientv3.NewWatcher(global.Etcd)
	defer watcher.Close()
	for true {
		watchRespChan := watcher.Watch(ctx, SDKeyPrefix+serviceNamePrefix, clientv3.WithPrefix(), clientv3.WithPrevKV())
		for watchResp := range watchRespChan {
			for _, event := range watchResp.Events {
				key := strings.Split(string(event.Kv.Key), "/")
				serviceName := key[1]
				serverId := key[2]
				serverAddr := string(event.Kv.Value)
				server := SDServer{serviceName: serviceName, serverId: serverId, ServerAddr: serverAddr}
				prevAddr := SDDefaultPrevAddr
				if event.PrevKv != nil {
					prevAddr = string(event.PrevKv.Value)
				}

				switch event.Type {
				case mvccpb.PUT:
					var service *SDService
					if tmp, ok := sd.cache.Load(serviceName); !ok {
						service = &SDService{
							List:           make(map[string]SDServer),
							index:          0,
							consistentHash: hash.NewConsistentHash(),
						}
						service.consistentHash.Add(serverId)
					} else {
						service = tmp.(*SDService)
					}
					service.List[serverId] = server
					sd.cache.Store(serviceName, service)
					if prevAddr != SDDefaultPrevAddr {
						global.Log.Infof("下线服务器 %s/%s:%s，上线服务器 %s/%s:%s\n", serviceName, serverId, prevAddr, serviceName, serverId, serverAddr)
					} else {
						global.Log.Infof("上线服务器 %s/%s:%s\n", serviceName, serverId, serverAddr)
					}
				case mvccpb.DELETE:
					var service *SDService
					if tmp, ok := sd.cache.Load(serviceName); ok {
						service = tmp.(*SDService)
						delete(service.List, serverId)
						service.consistentHash.Remove(serverId)
						sd.cache.Store(serviceName, service)
						global.Log.Debugf("下线服务器 %s/%s:%s\n", serviceName, serverId, prevAddr)
					}
				}
			}
		}
	}
}

// GetAll 获取所有服务、服务器信息
func (sd *serviceDiscovery) GetAll() map[string]*SDService {
	cache := make(map[string]*SDService)
	sd.cache.Range(func(key, value any) bool {
		cache[key.(string)] = value.(*SDService)
		return true
	})
	return cache
}

func (sd *serviceDiscovery) GetService(serviceName string) *SDService {
	if tmp, ok := sd.cache.Load(serviceName); ok {
		return tmp.(*SDService)
	}
	return nil
}

// DiscoveryByConsistentHash 通过一致性哈希获取一个不是自身的服务器地址
func (sd *serviceDiscovery) DiscoveryByConsistentHash(serviceName string, v interface{}) string {
	if tmp, ok := sd.cache.Load(serviceName); ok {
		service := tmp.(*SDService)
		if serverId, ok := service.consistentHash.Get(v); ok {
			if serverId.(string) != strconv.FormatInt(config.C.App.Id, 10) {
				if server, ok := service.List[serverId.(string)]; ok {
					return server.ServerAddr
				}
			}
		}
	}
	return ""
}
