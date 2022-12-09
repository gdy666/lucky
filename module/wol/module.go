package wol

import (
	"github.com/gdy666/lucky/config"
	"github.com/sirupsen/logrus"
)

func Init(log *logrus.Logger) {
	WOLClientConfigureInit(log)
	deviceInit(log)
}

// deviceInit 暂时用于第三方物联网平台部分初始化
func deviceInit(log *logrus.Logger) {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	for i := range config.Configure.WOLDeviceList {
		config.Configure.WOLDeviceList[i].SetShutDownFunc(ExecShutDown)
		go config.Configure.WOLDeviceList[i].DianDengClientStart()
		go config.Configure.WOLDeviceList[i].BemfaClientStart()
	}

}
