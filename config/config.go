package config

import (
	"github.com/spf13/viper"
	"go-com/core/pg"
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
	App app
	Pg  pg.Config
}

type app struct {
	Id            int64
	DebugMode     bool
	IsDistributed bool
	Environment   int
	Prefix        string
	RuntimePath   string

	ApiAddr                       string
	WebApiAddr                    string
	KafkaToLog                    bool
	MaxKafkaConsumeWorkerNum      int32
	MaxDelayQueueConsumeWorkerNum int32

	GatewayAddr  string
	GatewayToken string
}

// 将配置参数格式化为内存数据结构
func decode() {
	RuntimePath = C.App.RuntimePath
	if RuntimePath == "" {
		RuntimePath = Root + "runtime"
	}
	C.App.RuntimePath = RuntimePath
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
}
