package config

import (
	"flag"
	"github.com/spf13/viper"
	"log"
	"os"
	"regexp"
	"strconv"
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
	Mysql       Mysql
	Pg          Postgresql
	Redis       Redis
	KafkaP      Kafka
	KafkaC      Kafka
	Es          Es
	Etcd        Etcd
	Oracle      Oracle
	Clickhouse  Clickhouse
	Nebula      Nebula
	RateLimit   RateLimit
	RateBreaker RateBreaker
	Dq          Dq
	Ftp         Ftp
}

type app struct {
	Id            int64
	DebugMode     bool
	IsDistributed bool
	Environment   int
	Prefix        string
	RuntimePath   string
	LogExpire     int

	AppAddr  string
	WebAddr  string
	GrpcAddr string

	GatewayToken        string `json:"-"`
	ManageToken         string `json:"-"`
	ServiceSupportToken string `json:"-"`
	GrpcToken           string `json:"-"`

	DbType int

	PublicPath         string
	MaxMultipartMemory int64
}

type DbConfig struct {
	Prefix          string // 表前缀
	MaxOpenConns    int    // 设置最大连接数
	MaxIdleConns    int    // 设置最大空闲连接数
	ConnMaxLifetime int    // 设置一个连接的最大存活时长，单位：秒
}

type Mysql struct {
	Addr     string
	User     string
	Password string
	Dbname   string
	DbConfig DbConfig
}

type Postgresql struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
	Schema   string
	DbConfig DbConfig
}

type Redis struct {
	Addr       string
	Username   string
	Password   string
	Db         int
	ClientName string
	PoolSize   int

	Type int // 1 单机 2 哨兵

	MasterName       string
	SentinelAddrs    []string
	SentinelUsername string
	SentinelPassword string
}

type Kafka struct {
	Servers          string
	Username         string
	Password         string
	Topic            string
	Group            string
	SecurityProtocol string
	SaslMechanisms   string

	LogExpire           int
	MaxConsumeWorkerNum int
}

type Es struct {
	Addr     []string
	User     string
	Password string
	Prefix   string
	DbConfig DbConfig
}

type Etcd struct {
	Addr          []string
	User          string
	Password      string
	CertFile      string
	KeyFile       string
	TrustedCAFile string
}

type Oracle struct {
	Host     string
	Port     int
	User     string
	Password string
	Service  string
	DbConfig DbConfig
}

type Clickhouse struct {
	Addr     string
	User     string
	Password string
	Dbname   string
	DbConfig DbConfig
}

type Nebula struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
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

type Dq struct {
	ConsumePeriod       int
	PersistPeriod       int
	MaxWorkerNum        int32
	CheckNoExist        bool
	CheckNoRunningExist bool
}

type Ftp struct {
	Addr     string
	User     string
	Password string
}

// 完善配置项
func complete() {
	initDefine() // 全局初始化

	if C.App.RuntimePath == "" {
		RuntimePath = Root + "runtime"
	}
	if C.App.PublicPath == "" {
		C.App.PublicPath = Root + "runtime/public"
	}

	C.App.MaxMultipartMemory *= MB
}

// Load 初始化配置
func Load() {
	/***** 通过启动指令配置 *****/

	var id int64
	var configFile string
	flag.StringVar(&configFile, "config", "", "配置文件config.yaml的全路径") // -config=/data/auth-fast/config/config.yaml
	flag.Int64Var(&id, "id", 0, "在当前服务下的唯一编号，每启动一个服务程序都要配置，最大99")
	flag.Parse()

	/***** 通过config.yaml配置文件加载配置项 *****/

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

	/***** 多种途径加载系统id，并验证 *****/

	if id > 0 {
		C.App.Id = id
	}

	// 从环境变量中获取，用于k8s的StatefulSet方式，可从pod名称（xx-0、xx-1）中提取ID
	idEnv := os.Getenv("AUTH_FAST_ID")
	if idEnv != "" {
		reg, _ := regexp.Compile(`\d+`)
		idString := reg.FindString(idEnv)
		if idString == "" {
			log.Fatal("环境变量的ID有误。")
		} else {
			id, _ := strconv.Atoi(idString)
			C.App.Id = int64(id + 1)
		}
	}

	// 验证id
	if C.App.Id > 99 || C.App.Id < 1 {
		log.Fatal("id值在1和99之间。")
	}

	/***** 完善配置项 *****/

	complete()
}
