package etcd

import (
	"crypto/tls"
	"go-com/core/logr"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var Etcd *clientv3.Client

type Config struct {
	Addr          []string
	User          string
	Password      string
	CertFile      string
	KeyFile       string
	TrustedCAFile string
}

func InitEtcd(cfg Config) {
	Etcd = NewEtcd(cfg)
}

func NewEtcd(cfg Config) *clientv3.Client {
	var err error
	var e *clientv3.Client
	var tlsConfig *tls.Config
	if cfg.CertFile != "" {
		tlsInfo := &transport.TLSInfo{
			CertFile:      cfg.CertFile,
			KeyFile:       cfg.KeyFile,
			TrustedCAFile: cfg.TrustedCAFile,
		}
		tlsConfig, err = tlsInfo.ClientConfig()
		if err != nil {
			logr.L.Fatal(err)
		}
	}
	if e, err = clientv3.New(clientv3.Config{
		Endpoints:   cfg.Addr,
		DialTimeout: time.Second * 5, // client 首次连接超时，后面不用管，会自动重连
		Username:    cfg.User,
		Password:    cfg.Password,
		TLS:         tlsConfig,
	}); err != nil {
		logr.L.Fatal(err)
	}
	return e
}
