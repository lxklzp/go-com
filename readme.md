# go-com 项目构建框架

## 支持功能
* gin
* gorm（curd，excel导出）
* 多配置文件（viper yaml）
* 服务注册发现（etcd）、负载均衡（轮询、一致性哈希）、令牌桶限流（time/rate）、断路器（sony/gobreaker）
* 雪花算法
* 延迟队列（container/heap，定时轮询消费、队列消息持久化）
* grpc、http、rpc
* 定时任务（robfig/cron/v3）
* 表达式引擎（expr-lang/expr）
* 反向代理（标准库ReverseProxy）
* 数据同步：pg2my
* 日志（sirupsen/logrus）
* 加密：rsa、https证书、aes、3des、md5
* 版本号
* 打包

## 支持的数据存储与交互类型
* mysql、postgresql、clickhouse、nebula、oracle
* kafka、rabbitmq、redis、etcd、elasticsearch
* excel、csv、zip、file
* tcp、upd、email、ftp
