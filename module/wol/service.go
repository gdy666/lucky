package wol

import (
	"fmt"
	"sync"
	"time"

	wolconf "github.com/gdy666/lucky/module/wol/conf"
	websocketcontroller "github.com/gdy666/lucky/thirdlib/gdylib/websocketController"
)

var clientControllerMap sync.Map

func ReceiveMsgFromWOLClient(c *websocketcontroller.Controller, msgBytes []byte) {

	rawMsg, err := UnPack(msgBytes)
	if err != nil {
		return
	}
	switch m := rawMsg.(type) {
	case *Login:
		go hanlderWOLDeviceClientLogin(m, c)
	default:
		fmt.Printf("未知消息类型\n")
	}
}

func SyncClientConfigureToClient(d *wolconf.WOLDevice) {

	wc, wcOk := clientControllerMap.Load(d.Key)
	if !wcOk {
		return
	}

	syncMsg := &SyncClientConfigure{}
	syncMsg.DeviceName = d.DeviceName
	syncMsg.BroadcastIP = d.BroadcastIPs[0]
	syncMsg.Mac = d.MacList[0]
	syncMsg.Port = d.Port
	syncMsg.Relay = d.Relay
	syncMsg.Repeat = d.Repeat
	syncMsg.UpdateTime = d.UpdateTime
	SendMessage(wc.(*websocketcontroller.Controller), syncMsg)
}

func WOLClientDisconnect(c *websocketcontroller.Controller) {

	mac, macOk := c.GetExtData("mac")
	if macOk {
		clientControllerMap.Delete(mac)
	}

	key, keyOk := c.GetExtData("key")
	if keyOk {
		clientControllerMap.Delete(key)
	}
}

func WOLClientConnected(c *websocketcontroller.Controller) {
	//fmt.Printf("WOLClientConnected \n")
}

func WakeUpFinishedCallback(reply bool, macList []string, broadcastIps []string, port, repeat int) {
	if !reply {
		return
	}
	msg := &ReplyWakeUp{MacList: macList, BroadcastIPs: broadcastIps, Port: port, Repeat: repeat}
	devicesList := GetWOLDeviceList()
	for _, d := range devicesList {
		if !d.Relay {
			continue
		}
		wc, wcOk := clientControllerMap.Load(d.Key)
		if !wcOk {
			continue
		}
		SendMessage(wc.(*websocketcontroller.Controller), msg)
	}
}

func hanlderWOLDeviceClientLogin(m *Login, c *websocketcontroller.Controller) {
	msg := &LoginResp{}
	defer func() {
		if msg.Ret != 0 {
			return
		}
		c.StoreExtData("mac", m.Mac)
		c.StoreExtData("key", m.Key)
		c.StoreExtData("deviceName", m.DeviceName)
		clientControllerMap.Store(m.Key, c)
		clientControllerMap.Store(m.Mac, c)
	}()

	nowTimeStamp := time.Now().Unix()

	if nowTimeStamp-m.ClientTimeStamp > 30 || m.ClientTimeStamp-nowTimeStamp > 30 {
		msg.Msg = "客户端和服务器端时间相差大于30秒,请先校对时间"
		msg.Ret = 1
		SendMessage(c, msg)
		return
	}

	if m.Token != GetWOLServiceConfigure().Server.Token {
		msg.Msg = "Token不匹配"
		msg.Ret = 1
		SendMessage(c, msg)
		return
	}
	d := GetWOLDeviceByKey(m.Key)
	if d == nil { //根据Key找不到记录
		d = GetWOLDeviceByMac(m.Mac)
		dev := &wolconf.WOLDevice{
			Key:          m.Key,
			DeviceName:   m.DeviceName,
			MacList:      []string{m.Mac},
			BroadcastIPs: []string{m.BroadcastIP},
			Port:         m.Port,
			Relay:        m.Relay,
			Repeat:       m.Repeat,
			UpdateTime:   m.UpdateTime,
		}

		if d == nil { //创建记录
			err := WOLDeviceListAdd(dev)
			if err == nil {
				msg.Ret = 0
				SendMessage(c, msg)

			} else {
				msg.Msg = fmt.Sprintf("添加唤醒设备记录出错:%s", err.Error())
				msg.Ret = 2
				SendMessage(c, msg)
			}
		} else { //修改相同Mac记录的设备未
			fmt.Printf("修改相同Mac记录的设备\n")
			err := WOLDeviceListReplace(d.Key, dev)
			if err == nil {
				msg.Ret = 0
				SendMessage(c, msg)
			} else {
				msg.Msg = fmt.Sprintf("替换唤醒设备记录出错:%s", err.Error())
				msg.Ret = 3
				SendMessage(c, msg)
			}
		}

		return
	}

	if d.UpdateTime == m.UpdateTime {
		//fmt.Printf("两端规则更新时间一致,不用同步\n")
		msg.Ret = 0
		SendMessage(c, msg)
		return
	}

	//fmt.Printf("d.UpdateTime:%d m.UpdateTime:%d\n", d.UpdateTime, m.UpdateTime)
	if d.UpdateTime > m.UpdateTime {

		msg.Ret = 0
		SendMessage(c, msg)
		//fmt.Printf("服务端配置较新,同步至客户端\n")
		syncMsg := &SyncClientConfigure{m.WOLClientConfigure}
		syncMsg.DeviceName = d.DeviceName
		syncMsg.BroadcastIP = d.BroadcastIPs[0]
		syncMsg.Mac = d.MacList[0]
		syncMsg.Port = d.Port
		syncMsg.Relay = d.Relay
		syncMsg.Repeat = d.Repeat
		syncMsg.UpdateTime = d.UpdateTime
		SendMessage(c, syncMsg)

	} else {
		//fmt.Printf("客户端配置较新,同步至服务器端")
		dev := &wolconf.WOLDevice{
			Key:                  m.Key,
			DeviceName:           m.DeviceName,
			MacList:              []string{m.Mac},
			BroadcastIPs:         []string{m.BroadcastIP},
			Port:                 m.Port,
			Relay:                m.Relay,
			Repeat:               m.Repeat,
			UpdateTime:           m.UpdateTime,
			IOT_DianDeng_Enable:  d.IOT_Bemfa_Enable,
			IOT_DianDeng_AUTHKEY: d.IOT_DianDeng_AUTHKEY,
			IOT_Bemfa_Enable:     d.IOT_Bemfa_Enable,
			IOT_Bemfa_SecretKey:  d.IOT_Bemfa_SecretKey,
			IOT_Bemfa_Topic:      d.IOT_Bemfa_Topic,
		}
		err := WOLDeviceListAlter(dev)
		if err == nil {
			msg.Ret = 0
			SendMessage(c, msg)
		} else {
			msg.Ret = 3
			msg.Msg = fmt.Sprintf("同步客户端配置到服务端出错:%s", err.Error())
		}
	}

}

func ExecShutDown(d *wolconf.WOLDevice) int {
	successCount := 0
	msg := &ShutDown{}
	for _, mac := range d.MacList {
		cc, ccOk := clientControllerMap.Load(mac)
		if !ccOk {
			continue
		}
		SendMessage(cc.(*websocketcontroller.Controller), msg)
		successCount++
	}
	return successCount
}
