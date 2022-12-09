package wol

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	wolconf "github.com/gdy666/lucky/module/wol/conf"
	websocketcontroller "github.com/gdy666/lucky/thirdlib/gdylib/websocketController"
	"github.com/sirupsen/logrus"
)

var wolClient *websocketcontroller.Controller
var wolClientMu sync.Mutex

var clientState string
var clientStateMsg string

func init() {
	wolconf.SetClientInitFunc(WOLClientInit)
	wolconf.SetClientDisconnectFunc(ClientDisconnect)
}

func GetClientState() string {
	return clientState
}

func GetClientStateMsg() string {
	return clientStateMsg
}

func ClientDisconnect() {
	wolClientMu.Lock()
	defer wolClientMu.Unlock()
	if wolClient == nil {
		return
	}
	go wolClient.Disconnect()
}

func WOLClientInit(logger *logrus.Logger, c *wolconf.WOLClientConfigure) {

	if !c.Enable || c.ServerURL == "" {
		return
	}
	client := websocketcontroller.Controller{}
	client.ReceiveMessageCallback = receiveMessageCallback
	client.ClientDisconnectedCallback = clientStop
	client.ClientReadyCallback = clientReady
	client.Logs = logger
	client.SetConnectRetry(true)
	client.SetConnectRetryInterval(time.Second * 3)
	client.SetReadDeadline(time.Second * 5)
	client.SetServerURL(fmt.Sprintf("%s/api/wol/service", c.ServerURL))
	client.ScureSkipVerify(!wolconf.GetHttpClientSecureVerify())
	client.SetSendMessageEncryptionFunc(SendMessageEncryptionFunc)
	client.SetReceiveMessageDecryptionFunc(ReceiveMessageDecryptionFunc)

	wolClientMu.Lock()
	defer wolClientMu.Unlock()
	wolClient = &client

	go wolClient.Connect()
}

func receiveMessageCallback(c *websocketcontroller.Controller, msgBytes []byte) {
	rawMsg, err := UnPack(msgBytes)
	if err != nil {
		return
	}

	switch m := rawMsg.(type) {
	case *LoginResp:
		go handlerLoginResp(m, c)
	case *SyncClientConfigure:
		go synsClientConfigureToClient(m)
	case *ReplyWakeUp:
		//fmt.Printf("中继唤醒包:%v\n", m)
		c.Logs.Infof("中继唤醒包:%v\n", m)
		go wolconf.WakeOnLan(false, m.MacList, m.BroadcastIPs, m.Port, m.Repeat, nil)
	case *ShutDown:

		go func() {
			conf := GetWOLServiceConfigure()
			c.Logs.Infof("执行关机指令:%s\n", conf.Client.PowerOffCMD)
			//<-time.After(time.Second * 1)
			cmd := strings.Split(conf.Client.PowerOffCMD, " ")

			var output []byte
			var err error
			//fileutils.OpenProgramOrFile(cmd)
			if len(cmd) == 1 {
				output, err = exec.Command(cmd[0], []string{}...).Output()
			} else {
				output, err = exec.Command(cmd[0], cmd[1:]...).Output()
			}

			if err != nil {
				c.Logs.Errorf("执行关机指令[%s]出错:%s:%s", conf.Client.PowerOffCMD, string(output), err.Error())
			}

		}()
	default:
	}
}

func clientStop(cc *websocketcontroller.Controller) {
	//fmt.Printf("客户端已断开服务器连接\n")
	//clientState = "服务端已断开连接"
	//clientStateMsg = ""
}

func clientReady(cc *websocketcontroller.Controller) {
	cc.Logs.Info("WOL 客户端已连接上服务端,发送登录消息...")
	logMsg := &Login{}
	logMsg.WOLClientConfigure = wolconf.GetWOLServiceConfigure().Client
	logMsg.ClientTimeStamp = time.Now().Unix()
	SendMessage(cc, logMsg)
}

func handlerLoginResp(m *LoginResp, c *websocketcontroller.Controller) {
	if m.Ret != 0 {
		c.Disconnect()
		c.Logs.Error("WOl 服务端登录失败:%s", m.Msg)
		clientState = "服务端登录失败"
		clientStateMsg = m.Msg
		return
	}
	c.Logs.Info("WOL服务端登录成功")
	clientState = "服务端登录成功"
	clientStateMsg = ""

}

func synsClientConfigureToClient(m *SyncClientConfigure) {
	//fmt.Printf("处理来自服务端同步配置\n")
	conf := GetWOLServiceConfigure()
	conf.Client.Relay = m.Relay
	conf.Client.DeviceName = m.DeviceName
	conf.Client.Mac = m.Mac
	conf.Client.BroadcastIP = m.BroadcastIP
	conf.Client.Port = m.Port
	conf.Client.Repeat = m.Repeat
	conf.Client.UpdateTime = m.UpdateTime
	AlterWOLClientConfigure(&conf, logger, false)
}
