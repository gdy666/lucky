package blinker

import "fmt"

const (
	VA_TYPE_LIGHT        = "light"
	VA_TYPE_OUTLET       = "outlet"
	VA_TYPE_MULTI_OUTLET = "multi_outlet"
	VA_TYPE_SENSOR       = "sensor"
	VA_TYPE_FAN          = "fan"
	VA_TYPE_AIRCONDITION = "aircondition"
)

type VoiceAssistant struct {
	DeviceType string //语言助手类型 (设备类型).
	VAType     string //语言助手类型  MIOT AliGenie DuerOS
	Device     *BlinkerDevice
	topic      string
}

func (v *VoiceAssistant) GetSKey() string {
	switch v.VAType {
	case "MIOT":
		return "miType"
	case "AliGenie":
		return "aliType"
	case "DuerOS":
		return "duerType"
	default:
		return ""
	}
}

func (v *VoiceAssistant) PowerChangeReply(msgid, st string) {
	state := "off"

	if st == "true" {
		state = "on"
	}

	// if v.VAType == "MIOT" {
	// 	if state == "on" {
	// 		state = "true"
	// 	} else {
	// 		state = "false"
	// 	}
	// }

	data := map[string]string{"pState": state}
	v.Device.SendMessage("vAssistant", v.GetToDevice(), msgid, data)
}

func (v *VoiceAssistant) QueryDeviceState(msgid string) {
	state := v.Device.state
	// if v.VAType == "MIOT" {
	// 	if state == "on" {
	// 		state = "true"
	// 	} else {
	// 		state = "false"
	// 	}
	// }
	data := map[string]string{"pState": state}
	v.Device.SendMessage("vAssistant", v.GetToDevice(), msgid, data)
}

func (v *VoiceAssistant) GetToDevice() string {
	// if v.Device.DetailInfo.Broker == "blinker" {
	// 	return "ServerReceiver"
	// }
	return v.topic
}

func CreateVoiceAssistant(deviceType, vaType string) *VoiceAssistant {
	switch vaType {
	case "MIOT":
		return &VoiceAssistant{DeviceType: deviceType, VAType: vaType, topic: fmt.Sprintf("%s_r", vaType)}
	case "AliGenie":
		return &VoiceAssistant{DeviceType: deviceType, VAType: vaType, topic: fmt.Sprintf("%s_r", vaType)}
	case "DuerOS":
		{
			newDeviceType := ""
			switch deviceType {
			case VA_TYPE_LIGHT:
				newDeviceType = "LIGHT"
			case VA_TYPE_OUTLET:
				newDeviceType = "SOCKET"
			case VA_TYPE_MULTI_OUTLET:
				newDeviceType = "MULTI_SOCKET"
			case VA_TYPE_SENSOR:
				newDeviceType = "AIR_MONITOR"
			default:
			}
			if newDeviceType == "" {
				return nil
			}
			return &VoiceAssistant{DeviceType: newDeviceType, VAType: vaType, topic: fmt.Sprintf("%s_r", vaType)}
		}
	default:
		return nil
	}
}
