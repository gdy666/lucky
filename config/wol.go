package config

import (
	"fmt"

	"github.com/gdy666/lucky/thirdlib/gdylib/netinterfaces"
	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
	"github.com/gdy666/lucky/thirdlib/go-wol"
)

type WOLDevice struct {
	Key          string
	DeviceName   string
	MacList      []string
	BroadcastIPs []string
	Port         int
	Relay        bool //交给中继设备发送
	Repeat       int  //重复发送次数
}

func (d *WOLDevice) WakeUp() error {
	return WakeOnLan(d.MacList, d.BroadcastIPs, d.Port, d.Repeat)
}

func WakeOnLan(macList []string, broadcastIps []string, port, repeat int) (err error) {
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
				wol.WakeUpRepeat(mac, bcst, "", port, repeat)
			}

		}
		return
	}

	for _, bcst := range globalBroadcastList {
		matchCount++
		for _, mac := range macList {
			wol.WakeUpRepeat(mac, bcst, "", port, repeat)
		}
	}

	return
}

//----------------------------------------

func GetWOLDeviceByKey(key string) *WOLDevice {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	index := -1

	for i := range programConfigure.WOLDeviceList {
		if programConfigure.WOLDeviceList[i].Key == key {
			index = i
			break
		}
	}

	if index < 0 {
		return nil
	}
	device := programConfigure.WOLDeviceList[index]
	return &device

}

func GetWOLDeviceList() []WOLDevice {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()
	var res []WOLDevice
	if programConfigure == nil {
		return res
	}

	for i := range programConfigure.WOLDeviceList {
		res = append(res, programConfigure.WOLDeviceList[i])
	}
	return res
}

func WOLDeviceListAdd(d *WOLDevice) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()

	if d.Key == "" {
		d.Key = stringsp.GetRandomString(8)
	}
	programConfigure.WOLDeviceList = append(programConfigure.WOLDeviceList, *d)
	return Save()
}

func WOLDeviceListAlter(d *WOLDevice) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	index := -1
	for i := range programConfigure.WOLDeviceList {
		if programConfigure.WOLDeviceList[i].Key == d.Key {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("key:%s 不存在", d.Key)
	}
	programConfigure.WOLDeviceList[index] = *d
	return Save()
}

func WOLDeviceListDelete(key string) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	deleteIndex := -1

	for i := range programConfigure.WOLDeviceList {
		if programConfigure.WOLDeviceList[i].Key == key {
			deleteIndex = i
			break
		}
	}

	if deleteIndex < 0 {
		return fmt.Errorf("key:%s 不存在", key)
	}
	programConfigure.WOLDeviceList = DeleteWOLDeviceListslice(programConfigure.WOLDeviceList, deleteIndex)
	return Save()
}

func DeleteWOLDeviceListslice(a []WOLDevice, deleteIndex int) []WOLDevice {
	j := 0
	for i := range a {
		if i != deleteIndex {
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}
