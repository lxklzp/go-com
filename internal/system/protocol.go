package system

import (
	"github.com/spf13/viper"
	"go-com/config"
	"log"
)

var ShiLian shiLian

type shiLian struct {
	DigitalLifeIndexTimeType map[int]string

	FtpDeliveryBindingChannel            map[string]string
	FtpDeliveryIsAi                      map[string]string
	FtpDeliveryIsCloud                   map[string]string
	FtpDeliveryIsNetworkQuality          map[string]string
	FtpDeliveryIsCloudvideoUploadQuality map[string]string
	FtpDeliveryIsAdmin                   map[string]string
	FtpDeliveryDeviceSource              map[string]string
	FtpDeliveryIsGiven                   map[string]string
	FtpDeliveryRelyTree                  map[string]string
	CloudvideoUploadBindSource           map[string]string
	CloudvideoUploadIndustryFlag         map[string]string
	DeviceOfflineQualityBrdSource        map[string]string
	DeviceOfflineQualityAccessProtocol   map[string]string
	ShiLianQualityCity                   []string
	DeviceInfoCityGetCount               int
	BandwidthMsgLevel                    map[string]string
	DeviceOnlineStatus                   map[int]string
}

func (p *shiLian) Init() {
	// 载入配置
	v := viper.New()
	configFile := config.ConfigPath + "/shi_lian.yaml"
	v.SetConfigFile(configFile)
	viper.AddConfigPath(config.ConfigPath)
	if err := v.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	if err := v.Unmarshal(&p); err != nil {
		log.Fatal(err)
	}
}
