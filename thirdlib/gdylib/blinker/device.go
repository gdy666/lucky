package blinker

import (
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/buger/jsonparser"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	HOST                = "https://iot.diandeng.tech"
	API_AUTH            = HOST + "/api/v1/user/device/diy/auth"
	API_HEARTBEAT       = HOST + "/api/v1/user/device/heartbeat"
	API_VOICE_ASSISTANT = HOST + "/api/v1/user/device/voice_assistant"
)

type BlinkerDevice struct {
	authKey string

	subTopic    string
	pubTopic    string
	exasubTopic string //aliyun特有
	exapubTopic string //aliyun特有

	client        MQTT.Client
	DetailInfo    BlinkerDetailInfo
	heartBeatChan chan uint8
	hbmu          sync.Mutex
	preSendTime   time.Time
	sendMsgChan   chan message
	//
	state string

	voiceAssistants map[string]*VoiceAssistant
}

type message struct {
	TargetType string
	Device     string
	MessageID  string
	Msg        any
}

func CreateBlinkerDevice(ak string) *BlinkerDevice {
	d := &BlinkerDevice{authKey: ak}
	d.voiceAssistants = make(map[string]*VoiceAssistant)
	d.state = "on"
	return d
}

func (d *BlinkerDevice) AddVoiceAssistant(v *VoiceAssistant) {
	v.Device = d
	d.voiceAssistants[v.VAType] = v
}

func (d *BlinkerDevice) SyncAssistants() error {
	for _, v := range d.voiceAssistants {
		skey := v.GetSKey()
		dataMap := make(map[string]string)
		dataMap["token"] = d.DetailInfo.IotToken
		dataMap[skey] = v.DeviceType

		dataBytes, _ := json.Marshal(dataMap)

		resp, err := http.Post(API_VOICE_ASSISTANT, "application/json", strings.NewReader(string(dataBytes)))
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

func (d *BlinkerDevice) RunSenderMessageService() {
	for m := range d.sendMsgChan {
		t := time.Since(d.preSendTime) - time.Millisecond*1100
		if t < 0 {
			//log.Printf("太快,睡眠一下:%d\n", -t)
			<-time.After(-t)
		}
		d.sendMessage(m.TargetType, m.Device, m.MessageID, m.Msg)
	}

}

func (d *BlinkerDevice) RunHeartBearTimer() {
	if !d.hbmu.TryLock() {
		return
	}
	defer d.hbmu.Unlock()
	log.Printf("开始心跳...\n")
	d.heartBeatChan <- uint8(1)
	for range d.heartBeatChan {
		d.heartBeat()
		<-time.After(time.Second * 599)
		d.heartBeatChan <- uint8(1)
	}
	log.Printf("心跳中止...\n")
}

func (d *BlinkerDevice) Init() error {
	apiurl := fmt.Sprintf("%s?authKey=%s", API_AUTH, d.authKey)
	resp, err := http.Get(apiurl)
	if err != nil {
		return fmt.Errorf("device init http.Get err:%s", err.Error())
	}

	var infoRes BlinkerInfoRes
	err = GetAndParseJSONResponseFromHttpResponse(resp, &infoRes)
	if err != nil {
		return fmt.Errorf("parse DeviceInfo resp err:%s", err.Error())
	}

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

func (d *BlinkerDevice) Login() error {
	opts := MQTT.NewClientOptions()

	brokeyURL := fmt.Sprintf("%s:%s", d.DetailInfo.Host, d.DetailInfo.Port)

	//brokeyURL := fmt.Sprintf("tcp://broker.diandeng.tech:%s", d.DetailInfo.Port)
	opts.AddBroker(brokeyURL)
	opts.SetClientID(d.DetailInfo.DeviceName)
	opts.SetUsername(d.DetailInfo.IotID)
	opts.SetPassword(d.DetailInfo.IotToken)

	//opts.SetKeepAlive(time.Second * 3)
	//opts.WillRetained = true

	//choke := make(chan [2]string)
	// opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
	// 	//choke <- [2]string{msg.Topic(), string(msg.Payload())}
	// 	msg.Payload()
	// })

	opts.SetOnConnectHandler(func(c MQTT.Client) {
		log.Printf("连接成功!")
		d.client = c
		c.Subscribe(d.subTopic, byte(0), d.ReceiveMessageHandler)
		//c.Subscribe(d.exasubTopic, byte(0), d.ReceiveMessageHandler)
		d.heartBeatChan = make(chan uint8, 1)
		go d.RunHeartBearTimer()
		d.sendMsgChan = make(chan message, 8)
		go d.RunSenderMessageService()
	})

	opts.OnConnectionLost = func(c MQTT.Client, err error) {
		log.Printf("连接丢失:%s\n", err.Error())
		close(d.heartBeatChan)
		close(d.sendMsgChan)
		d.client = nil
	}

	//opts.

	client := MQTT.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("连接出错:%s", token.Error())
	}

	<-time.After(time.Second * 60000)

	return nil
}

func (d *BlinkerDevice) heartBeat() error {

	//hr := fmt.Sprintf("%s?deviceName=%s&key=%s&heartbeat=600", SERVER+HEARTBEAT_URL, d.DetailInfo.DeviceName, d.authKey)

	hr := fmt.Sprintf("%s?deviceName=%s&key=%s&heartbeat=600", API_HEARTBEAT, d.DetailInfo.DeviceName, d.authKey)
	resp, err := http.Get(hr)
	if err != nil {
		return fmt.Errorf("device init http.Get err:%s", err.Error())
	}

	respBytes, err := GetBytesFromHttpResponse(resp)
	if err != nil {
		return err
	}

	fmt.Printf("HearBeat:%s\n", string(respBytes))
	return nil
}

func (d *BlinkerDevice) ReceiveMessageHandler(c MQTT.Client, m MQTT.Message) {

	log.Printf("接收到MQTT消息:【%s】%s\n", m.Topic(), m.Payload())

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

func (d *BlinkerDevice) voiceAssistantMessageHandler(from string, msg []byte) {

	fmt.Printf("from:%s msg:%s\n", from, string(msg))

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

func (d *BlinkerDevice) powerChange(va *VoiceAssistant, msgId, state string) {
	d.state = state
	if va != nil {
		va.PowerChangeReply(msgId, state)
	}
}

func (d *BlinkerDevice) ownAppMessagehandler(msg []byte) {
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
			fmt.Printf(` "data", "get":Value:%s`, getValue)
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

func (d *BlinkerDevice) formatMess2assistant(targetType, toDevice, msgid string, data any) ([]byte, error) {
	m := mess2assistant{DeviceType: targetType, Data: data, FromDeivce: d.DetailInfo.DeviceName, ToDevice: toDevice, MessageID: msgid}
	rawBytes, err := json.Marshal(m)
	if err != nil {
		return []byte{}, err
	}

	str := base64.StdEncoding.EncodeToString(rawBytes)
	log.Printf("回复语音助手:%s\n", string(rawBytes))
	//fmt.Printf("base64:%s\n", str)

	return []byte(str), nil
	//return rawBytes, nil
}

func (d *BlinkerDevice) formatMess2Device(targetType, toDevice string, data any) ([]byte, error) {
	m := mess2device{DeviceType: targetType, Data: data, FromDeivce: d.DetailInfo.DeviceName, ToDevice: toDevice}
	return json.Marshal(m)
}

func (d *BlinkerDevice) SendMessage(targetType, todevice, msgid string, msg any) {
	m := message{Device: todevice, Msg: msg, TargetType: targetType, MessageID: msgid}
	d.sendMsgChan <- m
}

func (d *BlinkerDevice) sendMessage(targetType, todevice, msgid string, msg any) error {
	if d.client == nil {
		return fmt.Errorf("SendMessage error:client == nil")
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

	fmt.Printf("topic:%s\n", pubTopic)

	if token := d.client.Publish(pubTopic, 0, false, payload); token.Wait() && token.Error() != nil {
		fmt.Printf("Publish error:%s\n", token.Error())
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
	if len(bytes) > 0 {
		err = json.Unmarshal(bytes, &result)
		if err != nil {
			//log.Printf("请求接口解析json结果失败! ERROR: %s\n", err)
			return fmt.Errorf("GetAndParseJSONResponseFromHttpResponse 解析JSON结果出错：%s", err.Error())
		}
	}
	return nil
}
