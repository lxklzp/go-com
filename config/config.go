package config

import (
	"flag"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var Root string // 根目录

var RuntimePath string // 运行时的缓存文件目录

var C config // 配置项

const (
	EnvDev  = 1
	EnvProd = 2
)

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
	App      app
	Redis    redis
	Mysql    mysql
	Pgsql    pgsql
	Etcd     Etcd
	Nebula   nebula
	Kafka    kafka
	Rabbitmq rabbitmq
	Enum     enum
}

type app struct {
	Id                            int64
	DebugMode                     bool
	IsDistributed                 bool
	Environment                   int
	Prefix                        string
	RuntimePath                   string
	ApiAddr                       string
	WebApiAddr                    string
	KafkaToLog                    bool
	MaxKafkaConsumeWorkerNum      int32
	MaxDelayQueueConsumeWorkerNum int32

	GatewayAddr  string
	GatewayToken string
	AppApiAddr   string
}

type redis struct {
	Addr     string
	Password string
	Db       int
}

type mysql struct {
	Addr            string
	User            string
	Password        string
	Db              string
	ConnMaxLifetime time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
}

type pgsql struct {
	Host     string
	Port     string
	User     string
	Password string
	Dbname   string
}

type Etcd struct {
	Addr          []string
	User          string
	Password      string
	CertFile      string
	KeyFile       string
	TrustedCAFile string
}

type nebula struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
}

type kafka struct {
	Servers          string
	Username         string
	Password         string
	Topic            string
	Group            string
	SecurityProtocol string
	SaslMechanisms   string
}

type rabbitmq struct {
	Addr     string
	User     string
	Password string
}

type enum struct {
	City []string
}

var City map[int8]string
var CityIndex map[string]int8
var CitySlice []string
var CityCode map[string]string
var CityCodeIndex map[string]string

// 将配置参数格式化为内存数据结构
func decode() {
	RuntimePath = C.App.RuntimePath
	if RuntimePath == "" {
		RuntimePath = Root + "runtime"
	}
	C.App.RuntimePath = RuntimePath

	var row []string
	var index int
	City = make(map[int8]string)
	CityIndex = make(map[string]int8)
	CityCode = make(map[string]string)
	CityCodeIndex = make(map[string]string)

	for _, s := range C.Enum.City {
		row = strings.Split(s, ":")
		index, _ = strconv.Atoi(row[0])
		City[int8(index)] = row[1]
		CityIndex[row[1]] = int8(index)
		CitySlice = append(CitySlice, row[1])
		CityCode[row[2]] = row[1]
		CityCodeIndex[row[1]] = row[2]
	}
}

// Load 加载配置文件
func Load() {
	v := viper.New()
	configPath := Root + "config"
	configFile := configPath + "/config.yaml"
	v.SetConfigFile(configFile)
	viper.AddConfigPath(configPath)
	if err := v.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	if err := v.Unmarshal(&C); err != nil {
		log.Fatal(err)
	}

	decode()

	// 通过启动指令配置
	var id int64
	flag.Int64Var(&id, "id", 0, "在当前服务下的唯一编号，每启动一个服务程序都要配置，最大1023")
	flag.Parse()
	if id > 0 {
		C.App.Id = id
	}

	// 监控配置文件变化，用于热更新
	viper.OnConfigChange(func(e fsnotify.Event) {
		if err := viper.Unmarshal(&C); err != nil {
			log.Fatal(err)
		}
		decode()
	})
	viper.WatchConfig()
}
