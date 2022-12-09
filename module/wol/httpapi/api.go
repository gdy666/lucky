package wolhttpapi

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gdy666/lucky/module/service"
	"github.com/gdy666/lucky/module/wol"
	wolconf "github.com/gdy666/lucky/module/wol/conf"
	"github.com/gdy666/lucky/thirdlib/gdylib/netinterfaces"
	websocketcontroller "github.com/gdy666/lucky/thirdlib/gdylib/websocketController"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func SetLogger(l *logrus.Logger) {
	logger = l
}

// RegisterAPI 注册相关API
func RegisterAPI(r *gin.Engine, authorized *gin.RouterGroup) {
	authorized.POST("/api/wol/device", AddWOLDevice)
	authorized.GET("/api/wol/device/wakeup", WOLDeviceWakeUp)
	authorized.GET("/api/wol/device/shutdown", WOLDeviceShutDown)
	authorized.GET("/api/wol/devices", GetWOLDeviceList)
	authorized.PUT("/api/wol/device", AlterWOLDevice)
	authorized.DELETE("/api/wol/device", DeleteWOLDevice)
	authorized.GET("/api/wol/service/configure", WOLServiceConfigure)
	authorized.PUT("/api/wol/service/configure", WOLAlterServiceConfigure)
	authorized.GET("api/wol/service/getipv4interface", GetIPv4Interface)
	r.GET("/api/wol/service", WOLWebsocketHandler)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096 * 1,
	WriteBufferSize: 4096 * 1,
	CheckOrigin: func(r *http.Request) bool { //允许跨域
		return true
	},
}

func AddWOLDevice(c *gin.Context) {
	var requestObj wolconf.WOLDevice
	err := c.BindJSON(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "请求解析出错"})
		return
	}

	err = checkWolDevice(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": fmt.Sprintf("添加网络唤醒设备出错:%s", err.Error())})
		return
	}

	err = wol.WOLDeviceListAdd(&requestObj)
	if err != nil {
		log.Printf("config.WOLDeviceListAdd error:%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": fmt.Sprintf("添加网络唤醒设备出错!:%s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

type DeviceInfo struct {
	wolconf.WOLDevice
	State               string
	OnlineMacList       []string
	DianDengClientState string
	BemfaClientState    string
}

func GetWOLDeviceList(c *gin.Context) {

	rawlist := wol.GetWOLDeviceList()

	var list []DeviceInfo
	for i := range rawlist {
		state, onlineMacList := wol.GetDeviceStateDetail(&rawlist[i])
		d := DeviceInfo{}
		d.WOLDevice = rawlist[i]
		d.State = state
		d.OnlineMacList = onlineMacList
		d.DianDengClientState = rawlist[i].GetDianDengClientState()
		d.BemfaClientState = rawlist[i].GetBemfaClientState()
		list = append(list, d)
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0, "list": list})
}

func AlterWOLDevice(c *gin.Context) {
	var requestObj wolconf.WOLDevice
	err := c.BindJSON(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "请求解析出错"})
		return
	}

	err = checkWolDevice(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": fmt.Sprintf("修改网络唤醒设备出错:%s", err.Error())})
		return
	}

	err = wol.WOLDeviceListAlter(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("修改网络唤醒设备配置失败:%s", err.Error())})
		return
	}

	wol.SyncClientConfigureToClient(&requestObj)

	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func DeleteWOLDevice(c *gin.Context) {
	key := c.Query("key")
	err := wol.WOLDeviceListDelete(key)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("删除网络唤醒设备失败:%s", err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func WOLDeviceWakeUp(c *gin.Context) {
	key := c.Query("key")

	device := wol.GetWOLDeviceByKey(key)
	if device == nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "找不到Key对应的设备,唤醒失败"})
		return
	}
	err := device.WakeUp(wol.WakeUpFinishedCallback)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": "唤醒失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func WOLDeviceShutDown(c *gin.Context) {

	key := c.Query("key")
	device := wol.GetWOLDeviceByKey(key)
	if device == nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "找不到Key对应的设备,发送执行关机指令失败"})
		return
	}

	count := wol.ExecShutDown(device)
	if count == 0 {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": "没有设备在线,未能发送关机指令"})
		return
	}

	c.SecureJSON(http.StatusOK, gin.H{"ret": 0, "successCount": count})

}

func checkWolDevice(d *wolconf.WOLDevice) error {

	if len(d.MacList) <= 0 {
		return fmt.Errorf("网卡MAC不能为空")
	}

	if strings.HasPrefix(d.Key, "Client_") {
		if len(d.MacList) > 1 {
			return fmt.Errorf("与客户端关联的设备(由客户端连接同步)只能填写一个mac")
		}
		if len(d.BroadcastIPs) > 1 {
			return fmt.Errorf("与客户端关联的设备(由客户端连接同步)只能填写一个广播地址")
		}
	}

	if d.IOT_Bemfa_Enable {
		if strings.TrimSpace(d.IOT_Bemfa_SecretKey) == "" {
			return fmt.Errorf("巴法云私钥不能为空")
		}

		if strings.TrimSpace(d.IOT_Bemfa_Topic) == "" {
			return fmt.Errorf("巴法云主题不能为空")
		}

		if !strings.HasSuffix(d.IOT_Bemfa_Topic, "001") {
			return fmt.Errorf("巴法云主题需要以001结尾,表示插座类型")
		}

	}

	if d.IOT_DianDeng_Enable {
		if strings.TrimSpace(d.IOT_DianDeng_AUTHKEY) == "" {
			return fmt.Errorf("点灯科技设备密钥不能为空")
		}
	}

	if d.Port <= 0 || d.Port > 065535 {
		d.Port = 9
	}

	if d.Repeat <= 0 || d.Repeat > 10 {
		d.Repeat = 5
	}
	return nil
}

func WOLServiceConfigure(c *gin.Context) {
	conf := wol.GetWOLServiceConfigure()

	c.JSON(http.StatusOK, gin.H{
		"ret":            0,
		"configure":      conf,
		"ClientState":    wol.GetClientState(),
		"ClientStateMsg": wol.GetClientStateMsg(),
		"serviceStatus":  service.GetServiceState()})
}

func GetIPv4Interface(c *gin.Context) {
	interfacceList, err := netinterfaces.GetIPv4NetInterfaceInfoList()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("获取IPv4网络信息列表出错:%s", err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0, "list": interfacceList})
}

func WOLAlterServiceConfigure(c *gin.Context) {
	var conf wolconf.WOLServiceConfigure
	err := c.BindJSON(&conf)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "请求解析出错"})
		return
	}
	err = wol.AlterWOLClientConfigure(&conf, logger, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": fmt.Sprintf("保存配置出错:%s", err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0, "configure": conf})
}

func WOLWebsocketHandler(c *gin.Context) {
	if !wol.GetWOLServiceConfigure().Server.Enable {
		c.Abort()
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Errorf("WOLWebsocketHandler upgrader.Upgrade :error:%s", err.Error())
		c.Abort()
		return
	}

	wsc := new(websocketcontroller.Controller)
	wsc.Logs = logger
	wsc.ReceiveMessageCallback = wol.ReceiveMsgFromWOLClient
	wsc.ClientDisconnectedCallback = wol.WOLClientDisconnect
	wsc.ClientReadyCallback = wol.WOLClientConnected
	wsc.SetSendMessageEncryptionFunc(wol.SendMessageEncryptionFunc)
	wsc.SetReceiveMessageDecryptionFunc(wol.ReceiveMessageDecryptionFunc)
	wsc.ConnectReady(conn)

}
