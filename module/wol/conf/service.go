package wolconf

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type WOLServerConfigure struct {
	Enable bool
	Token  string
}

type WOLClientConfigure struct {
	Enable      bool   //开关
	ServerURL   string //服务器地址
	Token       string //验证token
	Relay       bool   //中继唤醒包
	Key         string
	DeviceName  string //设备名称
	Mac         string //网卡物理地址
	BroadcastIP string //广播地址
	Port        int    //端口
	Repeat      int    //重复次数
	PowerOffCMD string //关机指令
	UpdateTime  int64  //配置更新时间

}

type WOLServiceConfigure struct {
	Server WOLServerConfigure
	Client WOLClientConfigure
}

var serviceConfigure *WOLServiceConfigure
var serviceConfigureMu sync.RWMutex
var clientInitFunc func(logger *logrus.Logger, c *WOLClientConfigure)
var clientDisconnectFunc func()

func SetClientInitFunc(f func(logger *logrus.Logger, c *WOLClientConfigure)) {
	clientInitFunc = f
}

func SetClientDisconnectFunc(f func()) {
	clientDisconnectFunc = f
}

func GetWOLServiceConfigure() WOLServiceConfigure {
	serviceConfigureMu.RLock()
	defer serviceConfigureMu.RUnlock()
	conf := *serviceConfigure
	return conf
}

func StoreWOLServiceConfigure(con *WOLServiceConfigure) {
	serviceConfigureMu.Lock()
	defer serviceConfigureMu.Unlock()
	serviceConfigure = con
}

func (c *WOLClientConfigure) Init(logger *logrus.Logger) {
	if clientInitFunc != nil {
		clientInitFunc(logger, c)
	}
}

func (c *WOLClientConfigure) ClientDisconnect() {
	if clientDisconnectFunc != nil {
		clientDisconnectFunc()
	}
}
