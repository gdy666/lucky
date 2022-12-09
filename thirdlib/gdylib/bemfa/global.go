package bemfa

import (
	"fmt"
	"sync"
	"time"
)

var bemfaStore sync.Map
var bemfaStroeMu sync.Mutex

func GetBemfaDevice(secretKey string, httpClientSecureVerify bool, httpClientTimeout int) (*Device, error) {
	bemfaStroeMu.Lock()
	defer bemfaStroeMu.Unlock()
	device, deviceOk := bemfaStore.Load(secretKey)
	if deviceOk {
		d := device.(*Device)
		if d.OnLine() {
			return d, nil
		}

		if d.IsDisconnected() {

			d.Stop()

			err := d.Login()
			if err != nil {
				return nil, fmt.Errorf("bemfa Device login error:%s", err.Error())
			}

			i := 0
			for {
				<-time.After(time.Second * 200)
				i++
				if d.OnLine() {
					break
				}
				if i > 26 {
					break
				}
			}

			if d.OnLine() {
				return d, nil
			}

			return nil, fmt.Errorf("blinker drvice 连接服务器失败")
		}

		return device.(*Device), nil
	}
	d := CreateDevice(secretKey, httpClientSecureVerify, httpClientTimeout)

	err := d.Login()
	if err != nil {
		return nil, fmt.Errorf("bemfa Device Login error:%s", err.Error())
	}

	i := 0
	for {
		<-time.After(time.Millisecond * 100)
		i++
		if d.OnLine() {
			break
		}
		if i > 51 {
			break
		}
	}

	//fmt.Printf("在线\n")

	if d.OnLine() {
		bemfaStore.Store(secretKey, d)
		return d, nil
	}

	return nil, fmt.Errorf("bemfa drvice 连接服务器失败")
}

func UnRegisterPowerChangeCallback(d *Device, topic, key string) {
	bemfaStroeMu.Lock()
	defer bemfaStroeMu.Unlock()
	d.UnRegisterPowerChangeCallbackFunc(topic, key)

	if len(d.powerChangeCallbackMap) != 0 {
		return
	}

	d.Stop()
	bemfaStore.Delete(d.secretKey)

}
