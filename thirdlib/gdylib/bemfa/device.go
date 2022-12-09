package bemfa

import (
	"crypto/tls"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	bemfaHost = "bemfa.com"
	//mqttPort  = 9501
	mqttsPort = 9503
)

const (
	Disconnected uint32 = iota
	Connecting
	Reconnecting
	Connected
)

type Device struct {
	linkState              uint32
	secretKey              string
	extStroe               sync.Map
	secureVerify           bool
	httpClientTimeout      int
	powerChangeCallbackMap map[string]map[string]func(string)
	mu                     *sync.Mutex

	client MQTT.Client
	//clientMu sync.Mutex
}

func CreateDevice(secretKey string, secureVerify bool, httpClientTimeout int) *Device {
	d := &Device{secretKey: secretKey, secureVerify: secureVerify, httpClientTimeout: httpClientTimeout}
	d.powerChangeCallbackMap = make(map[string]map[string]func(string))
	d.mu = &sync.Mutex{}
	return d
}

func (d *Device) OnLine() bool {
	state := atomic.LoadUint32(&d.linkState)
	return state == Connected
}

func (d *Device) IsDisconnected() bool {
	state := atomic.LoadUint32(&d.linkState)
	return state == Disconnected
}

func (d *Device) ResigsterPowerChangeCallbackFunc(subTopic, key string, f func(string)) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.client == nil {
		return fmt.Errorf("client == nil")
	}
	if _, ok := d.powerChangeCallbackMap[subTopic]; ok { //已订阅
		d.powerChangeCallbackMap[subTopic][key] = f
	} else {
		d.powerChangeCallbackMap[subTopic] = make(map[string]func(string))
		d.powerChangeCallbackMap[subTopic][key] = f
		//fmt.Printf("订阅主题")
		d.client.Subscribe(subTopic, 1, d.ReceiveMessageHandler)
	}
	return nil
}

func (d *Device) UnRegisterPowerChangeCallbackFunc(subTopic, key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.powerChangeCallbackMap[subTopic]; !ok { //已订阅
		return
	}

	delete(d.powerChangeCallbackMap[subTopic], key)

	if len(d.powerChangeCallbackMap[subTopic]) <= 0 {
		delete(d.powerChangeCallbackMap, subTopic)
	}

}

func (d *Device) StoreExtData(key any, val any) {
	d.extStroe.Store(key, val)
}

func (d *Device) GetExtData(key any) (val any, ok bool) {
	val, ok = d.extStroe.Load(key)
	return
}

func (d *Device) Login() error {
	opts := MQTT.NewClientOptions()

	brokeyURL := fmt.Sprintf("mqtts://%s:%d", bemfaHost, mqttsPort)
	opts.AddBroker(brokeyURL)
	opts.SetClientID(d.secretKey)
	opts.SetUsername("")
	opts.SetPassword("")
	opts.SetConnectTimeout(time.Second * 2)
	opts.SetKeepAlive(time.Second * 30)
	opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: !d.secureVerify})
	opts.SetAutoReconnect(true)
	opts.ConnectRetryInterval = time.Second * 3

	opts.SetOnConnectHandler(func(c MQTT.Client) {
		log.Printf("巴法云 [%s]已连接\n", d.secretKey)
		d.mu.Lock()
		d.client = c
		d.mu.Unlock()
		atomic.StoreUint32(&d.linkState, Connected)
		d.SubTopicOption()
	})
	opts.OnConnectionLost = func(c MQTT.Client, err error) {
		log.Printf("巴法云 [%s]连接丢失:%s\n", d.secretKey, err.Error())
		atomic.StoreUint32(&d.linkState, Disconnected)

	}
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

func (d *Device) closeMQTTClient() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.client == nil {
		return
	}
	d.client.Disconnect(0)
	log.Printf("巴法云 [%s]主动关闭连接", d.secretKey)
	d.client = nil
}

func (d *Device) SubTopicOption() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.client == nil {
		return
	}
	topicList := []string{}

	for k := range d.powerChangeCallbackMap {
		topicList = append(topicList, k)
	}

	for _, t := range topicList {
		d.client.Subscribe(t, 1, d.ReceiveMessageHandler)
	}
}

func (d *Device) ReceiveMessageHandler(c MQTT.Client, m MQTT.Message) {
	//	fmt.Printf("接收到MQTT消息:\n[【%s】\n%s\n\n", m.Topic(), m.Payload())
	// switch string(m.Payload()) {
	// case "on":
	// 	c.Publish("switch001/up", 1, true, m.Payload())
	// case "off":
	// 	c.Publish("switch001/up", 1, true, m.Payload())
	// default:
	// }
	c.Publish(fmt.Sprintf("%s/up", m.Topic()), 1, true, m.Payload())
	go d.handlerReceivemessage(m.Topic(), string(m.Payload()))
}

func (d *Device) handlerReceivemessage(topic, msg string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.powerChangeCallbackMap[topic]; !ok {
		return
	}
	for _, f := range d.powerChangeCallbackMap[topic] {
		f(msg)
	}
}
