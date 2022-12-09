package wolconf

import (
	"fmt"
	"log"

	"github.com/gdy666/lucky/thirdlib/gdylib/bemfa"
	"github.com/gdy666/lucky/thirdlib/gdylib/blinker"
	"github.com/gdy666/lucky/thirdlib/gdylib/netinterfaces"
	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
	gowol "github.com/gdy666/lucky/thirdlib/go-wol"
)

var httpClientSecureVerify bool
var httpClientTimeout int

func GetHttpClientSecureVerify() bool {
	return httpClientSecureVerify
}

func SetHttpClientSecureVerify(b bool) {
	httpClientSecureVerify = b
}

func SetHttpClientTimeout(t int) {
	httpClientTimeout = t
}

type WOLDevice struct {
	Key          string
	DeviceName   string
	MacList      []string
	BroadcastIPs []string
	Port         int
	Relay        bool //交给中继设备发送
	Repeat       int  //重复发送次数
	//PowerOffCMD  string //关机指令
	UpdateTime int64 //配置更新时间

	IOT_DianDeng_Enable  bool //点灯科技开关
	IOT_DianDeng_AUTHKEY string
	dianDengClient       *blinker.Device
	dianDengClientMsg    string

	IOT_Bemfa_Enable    bool //巴法平台开关
	IOT_Bemfa_SecretKey string
	IOT_Bemfa_Topic     string
	bemfaClient         *bemfa.Device
	bemfaClientMsg      string

	shutDownFunc func(d *WOLDevice) int
}

func (d *WOLDevice) SetShutDownFunc(f func(d *WOLDevice) int) {
	d.shutDownFunc = f
}

func (d *WOLDevice) GetDianDengClientState() string {
	if d.dianDengClient == nil {
		if d.dianDengClientMsg != "" {
			return d.dianDengClientMsg
		}
		return "未设置"
	}

	if d.dianDengClient.OnLine() {
		return "已连接"
	}

	return "未连接"
}

func (d *WOLDevice) GetBemfaClientState() string {
	if d.bemfaClient == nil {
		if d.bemfaClientMsg != "" {
			return d.bemfaClientMsg
		}
		return "未设置"
	}

	if d.bemfaClient.OnLine() {
		return "已连接"
	}

	return "未连接"
}

func (d *WOLDevice) BemfaClientStart() {
	if !d.IOT_Bemfa_Enable || d.IOT_Bemfa_SecretKey == "" || d.IOT_Bemfa_Topic == "" {
		return
	}

	bemfaClient, err := bemfa.GetBemfaDevice(d.IOT_Bemfa_SecretKey, httpClientSecureVerify, httpClientTimeout)
	if err != nil {
		//fmt.Printf("FUCK:%s\n", err.Error())
		d.SetBemfaClientMsg(err.Error())
	}
	if bemfaClient != nil {
		d.SetBemfaClientMsg("")
		bemfaClient.ResigsterPowerChangeCallbackFunc(d.IOT_Bemfa_Topic, d.GetIdentKey(), d.powerChange)
		d.bemfaClient = bemfaClient
	}
}

func (d *WOLDevice) BemfaClientStop() {
	if d.bemfaClient == nil {
		return
	}
	bemfa.UnRegisterPowerChangeCallback(d.bemfaClient, d.IOT_Bemfa_Topic, d.GetIdentKey())
	d.bemfaClient = nil
}

func (d *WOLDevice) SetBemfaClientMsg(msg string) {
	d.bemfaClientMsg = msg
}

func (d *WOLDevice) GetBemfaClient() *bemfa.Device {
	return d.bemfaClient
}

func (d *WOLDevice) SetBemfaClient(dd *bemfa.Device) {
	d.bemfaClient = dd
}

//--------------------------

func (d *WOLDevice) DianDengClientStart() {
	if !d.IOT_DianDeng_Enable || d.IOT_DianDeng_AUTHKEY == "" {
		return
	}

	blinkerClient, err := blinker.GetBlinkerDevice(d.IOT_DianDeng_AUTHKEY, httpClientSecureVerify, httpClientTimeout)
	if err != nil {
		//fmt.Printf("FUCK:%s\n", err.Error())
		d.SetDianDengClientMsg(err.Error())
	}
	if blinkerClient != nil {
		d.SetDianDengClientMsg("")
		blinkerClient.RegisterPowerChangeCallbackFunc(d.GetIdentKey(), d.powerChange)
		d.dianDengClient = blinkerClient
	}
}

func (d *WOLDevice) DianDengClientStop() {
	if d.dianDengClient == nil {
		return
	}
	blinker.UnRegisterPowerChangeCallback(d.dianDengClient, d.GetIdentKey())
	d.dianDengClient = nil
}

func (d *WOLDevice) powerChange(state string) {
	log.Printf("WOLDevice 语音助手控制设备[%s]状态:%s\n", d.DeviceName, state)
	if state == "on" || state == "true" {
		d.WakeUp(nil)
	} else {
		if d.shutDownFunc != nil {
			d.shutDownFunc(d)
		}
	}
}

func (d *WOLDevice) SetDianDengClientMsg(msg string) {
	d.dianDengClientMsg = msg
}

func (d *WOLDevice) GetDianDengClient() *blinker.Device {
	return d.dianDengClient
}

func (d *WOLDevice) SetDianDengClient(dd *blinker.Device) {
	d.dianDengClient = dd
}

func (d *WOLDevice) GetIdentKey() string {
	return fmt.Sprintf("WOL:%s", d.Key)
}

func (d *WOLDevice) WakeUp(finishedCallback func(relay bool, macList []string, broadcastIps []string, port, repeat int)) error {
	return WakeOnLan(d.Relay, d.MacList, d.BroadcastIPs, d.Port, d.Repeat, finishedCallback)
}

func WakeOnLan(relay bool, macList []string, broadcastIps []string, port, repeat int,
	finishedCallback func(relay bool, macList []string, broadcastIps []string, port, repeat int),
) (err error) {
	defer func() {
		if finishedCallback != nil {
			finishedCallback(relay, macList, broadcastIps, port, repeat)
		}
	}()
	globalBroadcastList := netinterfaces.GetGlobalIPv4BroadcastList()
	matchCount := 0

	defer func() {
		if matchCount <= 0 {
			err = fmt.Errorf("找不到匹配的局域网广播IP,未能发送唤醒指令")
		}
	}()

	if len(broadcastIps) > 0 {
		for _, bcst := range broadcastIps {
			bcstOk := stringsp.StrIsInList(bcst, globalBroadcastList)
			if !bcstOk {
				continue
			}
			matchCount++
			for _, mac := range macList {
				gowol.WakeUpRepeat(mac, bcst, "", port, repeat)
			}

		}
		return
	}

	for _, bcst := range globalBroadcastList {
		matchCount++
		for _, mac := range macList {
			gowol.WakeUpRepeat(mac, bcst, "", port, repeat)
		}
	}

	return
}
