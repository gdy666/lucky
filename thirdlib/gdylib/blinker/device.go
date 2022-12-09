package blinker

import (
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/buger/jsonparser"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
)

const (
	Disconnected uint32 = iota
	Connecting
	Reconnecting
	Connected
)

func init() {

}

const (
	HOST                = "https://iot.diandeng.tech"
	API_AUTH            = HOST + "/api/v1/user/device/diy/auth"
	API_HEARTBEAT       = HOST + "/api/v1/user/device/heartbeat"
	API_VOICE_ASSISTANT = HOST + "/api/v1/user/device/voice_assistant"
)

type Device struct {
	linkState   uint32
	authKey     string
	subTopic    string
	pubTopic    string
	exasubTopic string //aliyun特有
	exapubTopic string //aliyun特有

	client        MQTT.Client
	clientMu      sync.Mutex
	DetailInfo    BlinkerDetailInfo
	heartBeatChan chan uint8
	preSendTime   time.Time
	//sendMsgChan   chan message

	extStroe               sync.Map
	powerChangeCallbackMap sync.Map
	state                  bool
	queryStateFunc         func() bool

	voiceAssistants        map[string]*VoiceAssistant
	httpClientSecureVerify bool
	httpClientTimeout      int
	httpclient             *http.Client
}

type message struct {
	TargetType string
	Device     string
	MessageID  string
	Msg        any
}

func CreateDevice(ak string, httpClientSecureVerify bool, httpClientTimeout int) *Device {
	d := &Device{authKey: ak, httpClientSecureVerify: httpClientSecureVerify, httpClientTimeout: httpClientTimeout}
	d.voiceAssistants = make(map[string]*VoiceAssistant)
	d.httpclient, _ = httputils.CreateHttpClient(
		"tcp",
		"",
		!httpClientSecureVerify,
		"",
		"",
		"",
		"",
		time.Duration(httpClientTimeout)*time.Second)
	return d
}

func (d *Device) GetState() bool {
	if d.queryStateFunc != nil {
		return d.queryStateFunc()
	}
	return d.state
}
func (d *Device) SetQueryStateFunc(f func() bool) {
	d.queryStateFunc = f
}

func (d *Device) RegisterPowerChangeCallbackFunc(key string, cb func(string)) {
	d.powerChangeCallbackMap.Store(key, cb)
}

func (d *Device) UnRegisterPowerChangeCallbackFunc(key string) {
	d.powerChangeCallbackMap.Delete(key)
}

func (d *Device) AddVoiceAssistant(v *VoiceAssistant) {
	v.Device = d
	d.voiceAssistants[v.VAType] = v
}

func (d *Device) StoreExtData(key any, val any) {
	d.extStroe.Store(key, val)
}

func (d *Device) GetExtData(key any) (val any, ok bool) {
	val, ok = d.extStroe.Load(key)
	return
}

func (d *Device) SyncAssistants() error {
	for _, v := range d.voiceAssistants {
		skey := v.GetSKey()
		dataMap := make(map[string]string)
		dataMap["token"] = d.DetailInfo.IotToken
		dataMap[skey] = v.DeviceType

		dataBytes, _ := json.Marshal(dataMap)

		resp, err := d.httpclient.Post(API_VOICE_ASSISTANT, "application/json", strings.NewReader(string(dataBytes)))
		if err != nil {
			return err
		}
		_, err = GetBytesFromHttpResponse(resp)
		if err != nil {
			return err
		}
		//fmt.Printf("同步语音助手结果:%s\n", respBytes)
	}

	return nil
}

// func (d *Device) RunSenderMessageService() {
// 	for m := range d.sendMsgChan {
// 		t := time.Since(d.preSendTime) - time.Millisecond*1100
// 		if t < 0 {
// 			//log.Printf("太快,睡眠一下:%d\n", -t)
// 			<-time.After(-t)
// 		}
// 		d.sendMessage(m.TargetType, m.Device, m.MessageID, m.Msg)
// 	}

// }

func (d *Device) RunHeartBearTimer() {

	ticker := time.NewTicker(time.Second * 599)

	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case _, ok := <-d.heartBeatChan:
			{
				if !ok {
					return
				}
				d.heartBeat()

			}
		case <-ticker.C:
			d.heartBeatChan <- uint8(1)
		}
	}

}

func (d *Device) Init() error {
	apiurl := fmt.Sprintf("%s?authKey=%s", API_AUTH, d.authKey)
	resp, err := d.httpclient.Get(apiurl)
	if err != nil {
		return fmt.Errorf("device init httpclient.Get err:%s", err.Error())
	}

	var infoRes BlinkerInfoRes

	//jsonparser.Get()

	respBytes, err := GetBytesFromHttpResponse(resp)
	if err != nil {
		return fmt.Errorf("GetBytesFromHttpResponse error:%s", err.Error())
	}

	messageRet, _ := jsonparser.GetInt(respBytes, "message")
	if messageRet != 1000 {
		detailStr, _ := jsonparser.GetString(respBytes, "detail")
		return fmt.Errorf("%s", detailStr)
	}

	err = json.Unmarshal(respBytes, &infoRes)

	if err != nil {
		return fmt.Errorf("登录过程解析登录结果出错:\n%s\n%s", string(respBytes), err.Error())
	}

	// err = GetAndParseJSONResponseFromHttpResponse(resp, &infoRes)
	// if err != nil {
	// 	return fmt.Errorf("parse DeviceInfo resp err:%s", err.Error())
	// }

	d.DetailInfo = infoRes.Detail

	err = d.SyncAssistants()
	if err != nil {
		return err
	}

	// if d.DetailInfo.Broker == "blinker" {
	// 	d.subTopic = fmt.Sprintf("/device/%s/r", d.DetailInfo.DeviceName)
	// 	d.pubTopic = fmt.Sprintf("/device/%s/s", d.DetailInfo.DeviceName)
	// 	return nil
	// }

	// if d.DetailInfo.Broker == "aliyun" {
	// 	d.subTopic = fmt.Sprintf("/%s/%s/r", d.DetailInfo.ProductKey, d.DetailInfo.DeviceName)
	// 	d.pubTopic = fmt.Sprintf("/%s/%s/s", d.DetailInfo.ProductKey, d.DetailInfo.DeviceName)
	// 	d.exasubTopic = "/device/ServerSender/r"
	// 	d.exapubTopic = "/device/ServerReceiver/s"
	// 	return nil
	// }

	d.subTopic = fmt.Sprintf("/device/%s/r", d.DetailInfo.DeviceName)
	d.pubTopic = fmt.Sprintf("/device/%s/s", d.DetailInfo.DeviceName)
	d.exasubTopic = "/device/ServerSender/r"
	d.exapubTopic = "/device/ServerReceiver/s"

	return nil
}

func (d *Device) closeMQTTClient() {
	d.clientMu.Lock()
	defer d.clientMu.Unlock()
	if d.client == nil {
		return
	}
	log.Printf("点灯科技 [%s]主动关闭连接", d.authKey)
	d.client.Disconnect(0)
	close(d.heartBeatChan)
	//close(d.sendMsgChan)
	d.client = nil
}

func (d *Device) OnLine() bool {
	state := atomic.LoadUint32(&d.linkState)
	return state == Connected
}

func (d *Device) IsDisconnected() bool {
	state := atomic.LoadUint32(&d.linkState)
	return state == Disconnected
}

func (d *Device) Login() error {
	opts := MQTT.NewClientOptions()

	brokeyURL := fmt.Sprintf("%s:%s", d.DetailInfo.Host, d.DetailInfo.Port)

	//brokeyURL := fmt.Sprintf("tcp://broker.diandeng.tech:%s", d.DetailInfo.Port)
	opts.AddBroker(brokeyURL)
	opts.SetClientID(d.DetailInfo.DeviceName)
	opts.SetUsername(d.DetailInfo.IotID)
	opts.SetPassword(d.DetailInfo.IotToken)
	opts.SetConnectTimeout(time.Second * 2)
	opts.SetKeepAlive(time.Second * 30)
	opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: !d.httpClientSecureVerify})
	opts.SetAutoReconnect(true)
	opts.ConnectRetryInterval = time.Second * 3

	opts.SetOnConnectHandler(func(c MQTT.Client) {
		atomic.StoreUint32(&d.linkState, Connected)
		log.Printf("点灯物联 [%s]已连接\n", d.authKey)
		d.clientMu.Lock()
		defer d.clientMu.Unlock()
		d.client = c
		c.Subscribe(d.subTopic, byte(0), d.ReceiveMessageHandler)
		d.heartBeatChan = make(chan uint8, 1)
		go d.RunHeartBearTimer()
		//d.sendMsgChan = make(chan message, 8)
		//go d.RunSenderMessageService()
	})

	//d.client.Disconnect()

	opts.SetConnectionLostHandler(func(c MQTT.Client, err error) {
		log.Printf("点灯物联 [%s]连接丢失:%s\n", d.authKey, err.Error())
		atomic.StoreUint32(&d.linkState, Disconnected)
		//d.closeMQTTClient()
		//fmt.Printf("SetConnectionLostHandler\n")
	})

	opts.SetReconnectingHandler(func(c MQTT.Client, opt *MQTT.ClientOptions) {
		atomic.StoreUint32(&d.linkState, Reconnecting)
	})
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("连接出错:%s", token.Error())
	}
	return nil
}

func (d *Device) Stop() {

	d.closeMQTTClient()
}

func (d *Device) heartBeat() error {

	//hr := fmt.Sprintf("%s?deviceName=%s&key=%s&heartbeat=600", SERVER+HEARTBEAT_URL, d.DetailInfo.DeviceName, d.authKey)

	hr := fmt.Sprintf("%s?deviceName=%s&key=%s&heartbeat=600", API_HEARTBEAT, d.DetailInfo.DeviceName, d.authKey)
	resp, err := d.httpclient.Get(hr)
	if err != nil {
		return fmt.Errorf("device init httpclient.Get err:%s", err.Error())
	}

	_, err = GetBytesFromHttpResponse(resp)
	if err != nil {
		return err
	}

	//fmt.Printf("HearBeat:%s\n", string(respBytes))
	return nil
}

func (d *Device) ReceiveMessageHandler(c MQTT.Client, m MQTT.Message) {

	//log.Printf("接收到MQTT消息:\n[【%s】\n%s\n\n", m.Topic(), m.Payload())

	if m.Topic() != d.subTopic {
		return
	}

	fromDevice, fromDeviceErr := jsonparser.GetString(m.Payload(), "fromDevice")
	if fromDeviceErr != nil || (fromDevice != d.DetailInfo.UUID && fromDevice != "ServerSender") {
		return
	}

	if fromDevice == d.DetailInfo.UUID {
		d.ownAppMessagehandler(m.Payload())
		return
	}

	from, fromErr := jsonparser.GetString(m.Payload(), "data", "from")
	if fromErr != nil {
		return
	}
	switch from {
	case "MIOT", "AliGenie", "DuerOS":
		d.voiceAssistantMessageHandler(from, m.Payload())
	default:
	}

}

func (d *Device) voiceAssistantMessageHandler(from string, msg []byte) {

	//fmt.Printf("from:%s msg:%s\n", from, string(msg))

	va, ok := d.voiceAssistants[from]
	if !ok {

		return
	}

	//fmt.Printf("voiceAssistantMessageHandler\t msg:[%s]\n", msg)
	messageId, messageIdErr := jsonparser.GetString(msg, "data", "messageId")
	if messageIdErr != nil {
		return
	}

	//jsonparser.GetString(msg, "data", "set")
	pstate, pstateErr := jsonparser.GetString(msg, "data", "set", "pState")
	if pstateErr == nil {
		d.powerChange(va, messageId, pstate)
		return
	}

	_, getKeyerr := jsonparser.GetString(msg, "data", "get")
	if getKeyerr == nil {
		va.QueryDeviceState(messageId)
	}
	//va.Power()

}

func (d *Device) powerChange(va *VoiceAssistant, msgId, state string) {
	if va != nil {
		va.PowerChangeReply(msgId, state)
	}

	if state == "true" || state == "on" {
		d.state = true
	} else {
		d.state = false
	}

	go func() {
		d.powerChangeCallbackMap.Range(func(key any, val any) bool {
			cb := val.(func(string))
			cb(state)
			return true
		})
	}()

}

func (d *Device) ownAppMessagehandler(msg []byte) {
	getValue, getKeyError := jsonparser.GetString(msg, "data", "get")
	if getKeyError == nil {
		switch getValue {
		case "state":
			d.SendMessage("OwnApp", d.DetailInfo.UUID, "", map[string]any{"state": "online"})
		case "timing":
			d.SendMessage("OwnApp", d.DetailInfo.UUID, "", map[string]any{"timing": map[string]any{"timing": []any{}}}) //{"timing":{"timing":[]}}
		case "countdown":
			d.SendMessage("OwnApp", d.DetailInfo.UUID, "", map[string]any{"countdown": "false"}) //`{ "countdown": false }`
		default:
			//fmt.Printf(` "data", "get":Value:%s`, getValue)
		}

		return
	}
}

type mess2device struct {
	DeviceType string `json:"deviceType"`
	Data       any    `json:"data"`
	FromDeivce string `json:"fromDevice"`
	ToDevice   string `json:"toDevice"`
}

type mess2assistant struct {
	DeviceType string `json:"deviceType"`
	Data       any    `json:"data"`
	FromDeivce string `json:"fromDevice"`
	ToDevice   string `json:"toDevice"`
	MessageID  string `json:"-"` //`json:"messageId"`
}

func (d *Device) formatMess2assistant(targetType, toDevice, msgid string, data any) ([]byte, error) {
	m := mess2assistant{DeviceType: targetType, Data: data, FromDeivce: d.DetailInfo.DeviceName, ToDevice: toDevice, MessageID: msgid}
	rawBytes, err := json.Marshal(m)
	if err != nil {
		return []byte{}, err
	}

	//str := base64.StdEncoding.EncodeToString(rawBytes)
	//log.Printf("回复语音助手:%s\n", string(rawBytes))
	//fmt.Printf("base64:%s\n", str)

	//return []byte(str), nil
	return rawBytes, nil
}

func (d *Device) formatMess2Device(targetType, toDevice string, data any) ([]byte, error) {
	m := mess2device{DeviceType: targetType, Data: data, FromDeivce: d.DetailInfo.DeviceName, ToDevice: toDevice}
	return json.Marshal(m)
}

func (d *Device) SendMessage(targetType, todevice, msgid string, msg any) {
	//m := message{Device: todevice, Msg: msg, TargetType: targetType, MessageID: msgid}
	//d.sendMsgChan <- m
	d.sendMessage(targetType, todevice, msgid, msg)
}

func (d *Device) sendMessage(targetType, todevice, msgid string, msg any) error {
	d.clientMu.Lock()
	defer d.clientMu.Unlock()
	if d.client == nil {
		return fmt.Errorf("d.Client == nil")
	}
	var pubTopic string
	var payload []byte
	var err error
	if targetType == "OwnApp" {
		pubTopic = d.pubTopic
		payload, err = d.formatMess2Device(targetType, todevice, msg)
		if err != nil {
			return err
		}
	} else if targetType == "vAssistant" {
		//pubTopic = "/device/ServerReceiver/s"
		//pubTopic = fmt.Sprintf("/sys/%s/%s/rrpc/response/%s", d.DetailInfo.ProductKey, d.DetailInfo.DeviceName, msgid)
		//pubTopic = fmt.Sprintf("%s", d.exapubTopic)
		pubTopic = d.pubTopic
		payload, err = d.formatMess2assistant(targetType, todevice, msgid, msg)
		if err != nil {
			return err
		}
	}

	// fmt.Printf("topic:%s\n", pubTopic)

	if token := d.client.Publish(pubTopic, 1, true, payload); token.Wait() && token.Error() != nil {
		//fmt.Printf("Publish error:%s\n", token.Error())
		return token.Error()
	}
	d.preSendTime = time.Now()

	return nil
}

//-----------------------

type BlinkerDetailInfo struct {
	Broker     string `json:"broker"`
	DeviceName string `json:"deviceName"`
	Host       string `json:"host"`
	IotID      string `json:"iotId"`
	IotToken   string `json:"iotToken"`
	Port       string `json:"port"`
	ProductKey string `json:"productKey"`
	UUID       string `json:"uuid"`
}

type BlinkerInfoRes struct {
	Message int               `json:"message"`
	Detail  BlinkerDetailInfo `json:"detail"`
}

// GetStringFromHttpResponse 从response获取
func GetBytesFromHttpResponse(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return []byte{}, fmt.Errorf("resp.Body = nil")
	}
	defer resp.Body.Close()
	var body []byte
	var err error
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return []byte{}, err
		}
		body, err = ioutil.ReadAll(reader)
		return body, err
	}
	body, err = ioutil.ReadAll(resp.Body)
	return body, err
}

func GetAndParseJSONResponseFromHttpResponse(resp *http.Response, result interface{}) error {
	bytes, err := GetBytesFromHttpResponse(resp)
	if err != nil {
		return fmt.Errorf("GetBytesFromHttpResponse err:%s", err.Error())
	}

	//	fmt.Printf("FUCK:\n%s\n", string(bytes))
	if len(bytes) > 0 {
		err = json.Unmarshal(bytes, &result)
		if err != nil {
			//log.Printf("请求接口解析json结果失败! ERROR: %s\n", err)
			return fmt.Errorf("GetAndParseJSONResponseFromHttpResponse 解析JSON结果出错：%s", err.Error())
		}
	}
	return nil
}
