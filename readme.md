# go-com 项目构建框架

## 支持功能
* gin
* gorm
* 配置（viper）
* 服务注册发现（etcd）、负载均衡（轮询、一致性哈希）、令牌桶限流（time/rate）、断路器（sony/gobreaker）
* 雪花算法
* 延迟队列（container/heap，定时轮询消费、队列消息持久化）
* grpc、http rpc
* 定时任务（robfig/cron/v3）
* 表达式引擎（expr-lang/expr）
* 日志（sirupsen/logrus）
* 打包

## 支持的数据类型
* mysql、postgresql、clickhouse、nebula
* kafka、rabbitmq、redis、etcd、elasticsearch
* excel、csv、zip、file
* tcp、upd、email
