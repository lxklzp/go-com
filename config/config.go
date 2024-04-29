package config

import (
	"flag"
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

var Root string // 根目录

var RuntimePath string // 运行时的缓存文件目录

var C config // 配置项

func init() {
	// 设置根目录
	Root, _ = os.Getwd()
	Root += "/"
	// 设置时区为东8区
	time.Local = time.FixedZone("CST", 8*3600)
	// 设置log
	log.SetFlags(log.Llongfile | log.Ldate | log.Ltime)
}

type config struct {
	App         app
	Postgresql  Postgresql
	Clickhouse  Clickhouse
	Kafka       Kafka
	Mysql       Mysql
	Redis       Redis
	Etcd        Etcd
	RateLimit   RateLimit
	RateBreaker RateBreaker
}

type app struct {
	Id            int64
	DebugMode     bool
	IsDistributed bool
	Environment   int
	Prefix        string
	RuntimePath   string

	PublicIp   string
	ApiAddr    string
	WebApiAddr string

	GatewayToken            string
	GatewayAddr             string
	PublicPath              string
	MaxMultipartMemory      int64
	GrpcAddr                string
	GrpcToken               string
	DelayQueueConsumePeriod int
	DelayQueuePersistPeriod int
}

type DbConfig struct {
	Prefix          string // 表前缀
	MaxOpenConns    int    // 设置最大连接数
	MaxIdleConns    int    // 设置最大空闲连接数
	ConnMaxLifetime int    // 设置一个连接的最大存活时长，单位：秒
}

type Postgresql struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
	DbConfig DbConfig
}

type Clickhouse struct {
	Addr     string
	User     string
	Password string
	Dbname   string
	DbConfig DbConfig
}

type Kafka struct {
	Servers          string
	Username         string
	Password         string
	Topic            string
	Group            string
	SecurityProtocol string
	SaslMechanisms   string

	IsLog               bool
	MaxConsumeWorkerNum int32
}

type Mysql struct {
	Addr     string
	User     string
	Password string
	Dbname   string
	DbConfig DbConfig
}

type Redis struct {
	Addr     string
	Password string
	Db       int
}

type Etcd struct {
	Addr          []string
	User          string
	Password      string
	CertFile      string
	KeyFile       string
	TrustedCAFile string
}

type RateLimit struct {
	Limit    int
	Burst    int
	Timeout  int32
	MaxStock int32
}

type RateBreaker struct {
	Interval          int
	OpenTimeout       int
	HafMaxRequests    uint32
	CloseMinRequests  uint32
	CloseErrorPercent uint32
}

// 将配置参数格式化为内存数据结构
func decode() {
	InitDefine()

	if C.App.RuntimePath == "" {
		RuntimePath = Root + "runtime"
	}
	if C.App.PublicPath == "" {
		C.App.PublicPath = Root + "runtime/public"
	}
}

// Load 加载配置文件
func Load() {
	// 通过启动指令配置
	var id int64
	var configFile string
	flag.StringVar(&configFile, "config", "", "配置文件config.yaml的全路径")
	flag.Int64Var(&id, "id", 0, "在当前服务下的唯一编号，每启动一个服务程序都要配置，最大1023")
	flag.Parse()

	v := viper.New()
	configPath := Root + "config"
	if configFile == "" {
		configFile = configPath + "/config.yaml"
	}
	v.SetConfigFile(configFile)
	viper.AddConfigPath(configPath)
	if err := v.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	if err := v.Unmarshal(&C); err != nil {
		log.Fatal(err)
	}

	decode()

	if id > 0 {
		C.App.Id = id
	}

	// 验证id
	if C.App.Id > 1023 || C.App.Id < 1 {
		log.Fatal("id值在1和1023之间")
	}
}
