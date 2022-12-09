package wol

import (
	"fmt"

	websocketcontroller "github.com/gdy666/lucky/thirdlib/gdylib/websocketController"
)

func SendMessage(c *websocketcontroller.Controller, msg any) error {
	msgBytes, err := Pack(msg)
	if err != nil {
		fmt.Printf("pack FUck:%s\n", err.Error())
		return err
	}
	c.SendMessage(msgBytes)
	return nil
}

func Pack(msg interface{}) ([]byte, error) {
	return msgCtl.Pack(msg)
}

func UnPack(bytes []byte) (msg Message, err error) {
	if len(bytes) <= 9 {
		err = fmt.Errorf("len(bytes) <= 9")
		return
	}
	return msgCtl.UnPack(bytes[0], bytes[9:])
}
