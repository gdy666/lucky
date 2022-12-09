package blinker

import (
	"fmt"
	"sync"
	"time"
)

var blinkerDeviceStore sync.Map
var blinkerdeviceStroeMu sync.Mutex

func GetBlinkerDevice(authKey string, httpClientSecureVerify bool, httpClientTimeout int) (*Device, error) {
	blinkerdeviceStroeMu.Lock()
	defer blinkerdeviceStroeMu.Unlock()
	device, deviceOk := blinkerDeviceStore.Load(authKey)
	if deviceOk {
		d := device.(*Device)
		if d.OnLine() {
			return d, nil
		}

		if d.IsDisconnected() {
			err := d.Init()
			if err != nil {
				return nil, fmt.Errorf("blinker Device init error:%s", err.Error())
			}

			d.Stop()

			err = d.Login()
			if err != nil {
				return nil, fmt.Errorf("blinker Device login error:%s", err.Error())
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
	d := CreateDevice(authKey, httpClientSecureVerify, httpClientTimeout)
	d.AddVoiceAssistant(CreateVoiceAssistant(VA_TYPE_OUTLET, "MIOT"))
	d.AddVoiceAssistant(CreateVoiceAssistant(VA_TYPE_OUTLET, "AliGenie"))
	d.AddVoiceAssistant(CreateVoiceAssistant(VA_TYPE_OUTLET, "DuerOS"))
	err := d.Init()
	if err != nil {
		return nil, fmt.Errorf("blinker Device init error:%s", err.Error())
	}

	err = d.Login()
	if err != nil {
		return nil, fmt.Errorf("blinker Device Login error:%s", err.Error())
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
		blinkerDeviceStore.Store(authKey, d)
		return d, nil
	}

	return nil, fmt.Errorf("blinker drvice 连接服务器失败")
}

// func RegisterPowerChangeCallback(authkey, key string, cb func(string)) (*Device, error) {
// 	d, err := GetBlinkerDevice(authkey)
// 	if err != nil {
// 		return nil, err
// 	}
// 	d.RegisterPowerChangeCallbackFunc(key, cb)
// 	return d, nil
// }

func UnRegisterPowerChangeCallback(d *Device, key string) {
	blinkerdeviceStroeMu.Lock()
	defer blinkerdeviceStroeMu.Unlock()
	d.UnRegisterPowerChangeCallbackFunc(key)

	isEmpty := true
	d.powerChangeCallbackMap.Range(func(key any, val any) bool {
		isEmpty = false
		return false
	})

	if isEmpty {
		d.Stop()
		blinkerDeviceStore.Delete(d.authKey)
	}
}
