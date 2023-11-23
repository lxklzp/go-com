package global

import (
	"crypto/tls"
	"go-com/config"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var Etcd *clientv3.Client

func InitEtcd() {
	var err error
	var tlsConfig *tls.Config
	cfg := config.C.Etcd
	if cfg.CertFile != "" {
		tlsInfo := &transport.TLSInfo{
			CertFile:      cfg.CertFile,
			KeyFile:       cfg.KeyFile,
			TrustedCAFile: cfg.TrustedCAFile,
		}
		tlsConfig, err = tlsInfo.ClientConfig()
		if err != nil {
			Log.Fatal(err)
		}
	}
	if Etcd, err = clientv3.New(clientv3.Config{
		Endpoints:   cfg.Addr,
		DialTimeout: time.Second * 5, // client 首次连接超时，后面不用管，会自动重连
		Username:    cfg.User,
		Password:    cfg.Password,
		TLS:         tlsConfig,
	}); err != nil {
		Log.Fatal(err)
	}
}
