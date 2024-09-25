package service

import (
	"context"
	"go-com/config"
	"go-com/core/hash"
	"go-com/core/logr"
	"go-com/core/tool"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// SDKeyPrefix etcd的key前缀
const SDKeyPrefix = "sd-go-com/"

const SDDefaultPrevAddr = "unknown_prev_addr"

// Discovery 服务发现 etcd的key结构：统一前缀/服务名称/服务器编号
type Discovery struct {
	etcd  *clientv3.Client
	cache sync.Map // map[string]SDService 服务名称->服务->服务器编号->服务器
}

func NewServiceDiscovery(e *clientv3.Client) *Discovery {
	return &Discovery{etcd: e}
}

// SDService 服务
type SDService struct {
	List           map[string]SDServer  // 服务器列表 服务器编号->服务器
	index          int64                // 服务发现次数，单调递增
	consistentHash *hash.ConsistentHash // 一致性哈希，存储服务器编号
}

// SDServer 服务器
type SDServer struct {
	serviceName string // 服务名称
	serverId    string // 服务器编号
	ServerAddr  string // 服务器地址
}

// Registry 服务注册，服务提供方
func (sd *Discovery) Registry(serviceName string, serverId string, ServerAddr string) {
	// 创建和声明一个租约，并且设置ttl（60秒）
	ctx := context.TODO()
	lease := clientv3.NewLease(sd.etcd)
	leaseGrant, err := lease.Grant(context.Background(), 60)
	if err != nil {
		logr.L.Fatal(err)
	}
	// 通过put记录SDServer，实现服务注册
	server := SDServer{
		serviceName: serviceName,
		serverId:    serverId,
		ServerAddr:  ServerAddr,
	}
	key := SDKeyPrefix + server.serviceName + "/" + server.serverId
	if _, err = sd.etcd.Put(ctx, key, server.ServerAddr, clientv3.WithLease(leaseGrant.ID)); err != nil {
		logr.L.Fatal(err)
	}

	defer func() {
		if err := recover(); err != nil {
			tool.ErrorStack(err)
		}
		lease.Close()
		sd.Registry(serviceName, serverId, ServerAddr)
	}()

	// 租约保活
	keepRespChan, err := lease.KeepAlive(ctx, leaseGrant.ID)
	if err != nil {
		logr.L.Fatal(err)
	}
	logr.L.Infof("etcd服务注册上线：%v", server)
	for range keepRespChan {
	}
	logr.L.Errorf("etcd服务注册掉线：%v", server)
}

// Watch 维护服务、服务器信息，服务使用方
func (sd *Discovery) Watch(serviceNamePrefix string) {
	ctx := context.TODO()

	// 通过get初始化列表
	getResp, err := sd.etcd.Get(ctx, SDKeyPrefix+serviceNamePrefix, clientv3.WithPrefix())
	if err != nil {
		logr.L.Fatal(err)
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
		logr.L.Infof("上线服务器：%v", server)
	}
	for k, v := range cache {
		sd.cache.Store(k, v)
	}

	// 监听前缀key的值变化
	watcher := clientv3.NewWatcher(sd.etcd)
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
						logr.L.Infof("下线服务器：{%s %s %s}，上线服务器：%v", serviceName, serverId, prevAddr, server)
					} else {
						logr.L.Infof("上线服务器：%v", server)
					}
				case mvccpb.DELETE:
					var service *SDService
					if tmp, ok := sd.cache.Load(serviceName); ok {
						service = tmp.(*SDService)
						delete(service.List, serverId)
						service.consistentHash.Remove(serverId)
						sd.cache.Store(serviceName, service)
						logr.L.Debugf("下线服务器：%v", server)
					}
				}
			}
		}
	}
}

// GetAll 获取所有服务、服务器信息
func (sd *Discovery) GetAll() map[string]*SDService {
	cache := make(map[string]*SDService)
	sd.cache.Range(func(key, value any) bool {
		cache[key.(string)] = value.(*SDService)
		return true
	})
	return cache
}

func (sd *Discovery) GetService(serviceName string) *SDService {
	if tmp, ok := sd.cache.Load(serviceName); ok {
		return tmp.(*SDService)
	}
	return nil
}

// DiscoveryByConsistentHash 通过一致性哈希获取一个不是自身的服务器地址
func (sd *Discovery) DiscoveryByConsistentHash(serviceName string, v interface{}) string {
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

// DiscoveryByRoundRobin 服务发现 负载均衡：round-robin（轮询）
func (sd *Discovery) DiscoveryByRoundRobin(serviceName string) string {
	if tmp, ok := sd.cache.Load(serviceName); ok {
		service := tmp.(*SDService)
		var list []string
		for _, server := range service.List {
			if server.serverId != strconv.FormatInt(config.C.App.Id, 10) {
				list = append(list, server.ServerAddr)
			}
		}
		if len(list) == 0 {
			return ""
		}
		sort.Strings(list)
		atomic.AddInt64(&service.index, 1)
		sd.cache.Store(serviceName, service)
		return list[service.index%int64(len(list))]
	}
	return ""
}
