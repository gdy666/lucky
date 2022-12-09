package wol

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gdy666/lucky/config"
	wolconf "github.com/gdy666/lucky/module/wol/conf"
	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
	websocketcontroller "github.com/gdy666/lucky/thirdlib/gdylib/websocketController"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func WOLClientConfigureInit(l *logrus.Logger) {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()

	wolconf.StoreWOLServiceConfigure(&config.Configure.WOLServiceConfigure)
	wolconf.SetHttpClientSecureVerify(config.Configure.BaseConfigure.HttpClientSecureVerify)
	wolconf.SetHttpClientTimeout(config.Configure.BaseConfigure.HttpClientTimeout)
	config.Configure.WOLServiceConfigure.Client.Init(l)
	logger = l
}

//----------------------------------------

func GetWOLServiceConfigure() wolconf.WOLServiceConfigure {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	conf := config.Configure.WOLServiceConfigure
	return conf
}

func AlterWOLClientConfigure(conf *wolconf.WOLServiceConfigure, logger *logrus.Logger, updateTimeOpt bool) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	if conf.Client.Key == "" {
		conf.Client.Key = fmt.Sprintf("Client_%s", stringsp.GetRandomString(16))
	}

	// if programConfigure.WOLClientConfigure.Client != nil {
	// 	go programConfigure.WOLClientConfigure.Client.Disconnect()
	// }
	if updateTimeOpt {
		config.Configure.WOLServiceConfigure.Client.ClientDisconnect()
	}

	conf.Client.ServerURL = strings.TrimSpace(conf.Client.ServerURL)

	conf.Client.ServerURL = handlerWOLServerURL(conf.Client.ServerURL)

	conf.Client.Token = strings.TrimSpace(conf.Client.Token)
	conf.Client.DeviceName = strings.TrimSpace(conf.Client.DeviceName)
	conf.Client.Mac = strings.TrimSpace(conf.Client.Mac)
	conf.Client.BroadcastIP = strings.TrimSpace(conf.Client.BroadcastIP)
	conf.Client.PowerOffCMD = strings.TrimSpace(conf.Client.PowerOffCMD)

	if err := CheckWOLServiceConfigure(conf); err != nil {
		return err
	}
	if updateTimeOpt {
		conf.Client.UpdateTime = time.Now().Unix()
	}

	config.Configure.WOLServiceConfigure = *conf
	if updateTimeOpt {
		config.Configure.WOLServiceConfigure.Client.Init(logger)
	}

	wolconf.StoreWOLServiceConfigure(&config.Configure.WOLServiceConfigure)
	return config.Save()
}

func handlerWOLServerURL(l string) string {
	if !strings.HasPrefix(l, "http") && !strings.HasPrefix(l, "ws") {
		l = "ws://" + l
	}

	u, e := url.Parse(l)
	if e != nil {
		return ""
	}
	scheme := ""
	switch u.Scheme {
	case "http":
		scheme = "ws"
	case "https":
		scheme = "wss"
	case "ws":
		scheme = "ws"
	case "wss":
		scheme = "wss"
	default:
		scheme = "未知协议"
	}
	l = fmt.Sprintf("%s://%s", scheme, u.Host)
	return l
}

func CheckWOLServiceConfigure(conf *wolconf.WOLServiceConfigure) error {

	if conf.Client.Enable && strings.TrimSpace(conf.Client.ServerURL) == "" {
		return fmt.Errorf("客户端设置 服务端地址不能为空")
	}

	if conf.Client.Enable && strings.TrimSpace(conf.Client.Token) == "" {
		return fmt.Errorf("客户端设置 Token不能为空")
	}

	if conf.Client.Enable && strings.TrimSpace(conf.Client.PowerOffCMD) == "" {
		return fmt.Errorf("客户端设置 关机指令不能为空")
	}

	//广播地址,设备名称,关机指令可以为空
	// if strings.TrimSpace(conf.DeviceName) == "" {
	// 	return fmt.Errorf("设备名称不能为空")
	// }

	if conf.Client.Enable && strings.TrimSpace(conf.Client.Mac) == "" {
		return fmt.Errorf("客户端设置 物理网卡地址不能为空")
	}

	if conf.Server.Enable && strings.TrimSpace(conf.Server.Token) == "" {
		return fmt.Errorf("服务端设置 Token不能为空")
	}

	return nil
}

func GetWOLDeviceByKey(key string) *wolconf.WOLDevice {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	index := -1

	for i := range config.Configure.WOLDeviceList {
		if config.Configure.WOLDeviceList[i].Key == key {
			index = i
			break
		}
	}

	if index < 0 {
		return nil
	}
	device := config.Configure.WOLDeviceList[index]
	return &device
}

func GetWOLDeviceByMac(mac string) *wolconf.WOLDevice {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	index := -1

	for i := range config.Configure.WOLDeviceList {
		if config.Configure.WOLDeviceList[i].MacList[0] == mac && len(config.Configure.WOLDeviceList[i].MacList) == 1 {
			index = i
			break
		}
	}

	if index < 0 {
		return nil
	}
	device := config.Configure.WOLDeviceList[index]
	return &device
}

func GetWOLDeviceList() []wolconf.WOLDevice {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	var res []wolconf.WOLDevice
	if config.Configure == nil {
		return res
	}

	for i := range config.Configure.WOLDeviceList {
		res = append(res, config.Configure.WOLDeviceList[i])
	}
	return res
}

func GetDeviceStateDetail(d *wolconf.WOLDevice) (state string, onlineMacList []string) {

	// clientControllerMap.Range(func(key any, val any) bool {
	// 	fmt.Printf("%v %v\n", key, val)
	// 	return true
	// })

	if strings.HasPrefix(d.Key, "Client_") {
		c, ok := clientControllerMap.Load(d.Key)
		if ok {
			deviceName := "未设置主机名"
			dn, ok := c.(*websocketcontroller.Controller).GetExtData("deviceName")
			if ok {
				deviceName = dn.(string)
			}
			remoteAddr := strings.Split(c.(*websocketcontroller.Controller).GetRemoteAddr(), ":")[0]
			onlineMacList = append(onlineMacList, fmt.Sprintf("主机名:%s [ %s ]  网卡物理地址:%s", deviceName, remoteAddr, d.MacList[0]))
			state = "在线"

		} else {
			state = "离线"
		}

		return
	}

	for i := range d.MacList {
		c, ok := clientControllerMap.Load(d.MacList[i])
		if ok {
			deviceName := "未设置主机名"
			dn, ok := c.(*websocketcontroller.Controller).GetExtData("deviceName")
			if ok {
				deviceName = dn.(string)
			}
			remoteAddr := strings.Split(c.(*websocketcontroller.Controller).GetRemoteAddr(), ":")[0]
			onlineMacList = append(onlineMacList, fmt.Sprintf("主机名:%s [ %s ] 网卡物理地址:%s", deviceName, remoteAddr, d.MacList[i]))
		}
	}

	if len(onlineMacList) == 0 { //离线
		// if len(d.MacList) == 1 {
		// 	state = "离线"
		// } else {
		// 	state = "全部离线"
		// }
		state = "未知"
		return
	}

	if len(onlineMacList) == len(d.MacList) {
		if len(d.MacList) == 1 {
			state = "在线"
		} else {
			state = "全部在线"
		}
		return
	}

	state = "部分在线"

	return
}

func WOLDeviceListAdd(d *wolconf.WOLDevice) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()

	if d.Key == "" {
		d.Key = stringsp.GetRandomString(8)
	}
	if d.UpdateTime == 0 {
		d.UpdateTime = time.Now().Unix()
	}

	d.SetShutDownFunc(ExecShutDown)
	config.Configure.WOLDeviceList = append(config.Configure.WOLDeviceList, *d)

	listLength := len(config.Configure.WOLDeviceList)
	go config.Configure.WOLDeviceList[listLength-1].DianDengClientStart()
	go config.Configure.WOLDeviceList[listLength-1].BemfaClientStart()

	return config.Save()
}

func WOLDeviceListAlter(d *wolconf.WOLDevice) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	index := -1
	for i := range config.Configure.WOLDeviceList {
		if config.Configure.WOLDeviceList[i].Key == d.Key {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("key:%s 不存在", d.Key)
	}
	d.UpdateTime = time.Now().Unix()

	config.Configure.WOLDeviceList[index].DianDengClientStop()
	config.Configure.WOLDeviceList[index].BemfaClientStop()
	//go d.DianDengClientStart()
	d.SetShutDownFunc(ExecShutDown)

	config.Configure.WOLDeviceList[index] = *d
	go config.Configure.WOLDeviceList[index].DianDengClientStart()
	go config.Configure.WOLDeviceList[index].BemfaClientStart()
	return config.Save()
}

func WOLDeviceListReplace(key string, d *wolconf.WOLDevice) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	index := -1
	for i := range config.Configure.WOLDeviceList {
		if config.Configure.WOLDeviceList[i].Key == key {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("key:%s 不存在", d.Key)
	}
	config.Configure.WOLDeviceList[index] = *d
	return config.Save()
}

func WOLDeviceListDelete(key string) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	deleteIndex := -1

	for i := range config.Configure.WOLDeviceList {
		if config.Configure.WOLDeviceList[i].Key == key {
			deleteIndex = i
			break
		}
	}

	if deleteIndex < 0 {
		return fmt.Errorf("key:%s 不存在", key)
	}
	config.Configure.WOLDeviceList[deleteIndex].DianDengClientStop()
	config.Configure.WOLDeviceList[deleteIndex].BemfaClientStop()
	config.Configure.WOLDeviceList = DeleteWOLDeviceListslice(config.Configure.WOLDeviceList, deleteIndex)
	return config.Save()
}

func DeleteWOLDeviceListslice(a []wolconf.WOLDevice, deleteIndex int) []wolconf.WOLDevice {
	j := 0
	for i := range a {
		if i != deleteIndex {
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}
