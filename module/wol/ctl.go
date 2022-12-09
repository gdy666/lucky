package wol

import (
	"encoding/base64"

	jsonMsg "github.com/gdy666/lucky/thirdlib/fatedier/golib/json"
	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
)

type Message = jsonMsg.Message

var (
	msgCtl *jsonMsg.MsgCtl
)
var msgkeyBytes = []byte("lucky666")

func init() {
	msgCtl = jsonMsg.NewMsgCtl()
	for typeByte, msg := range msgTypeMap {
		msgCtl.RegisterMsg(typeByte, msg)
	}
}

func SendMessageEncryptionFunc(messageBytesPtr []byte) ([]byte, error) {
	outs, _ := stringsp.DesEncrypt(messageBytesPtr, msgkeyBytes) //加密
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(outs)))
	base64.StdEncoding.Encode(buf, outs)
	return buf, nil
}

// receiveMessageDecryptionFunc 自定义接收消息解密函数
func ReceiveMessageDecryptionFunc(messageBytes []byte) ([]byte, error) {
	rawEncryptMsgBytes, err := base64.StdEncoding.DecodeString(string(messageBytes))
	if err != nil {
		return nil, err
	}
	rawMsgBytes, err := stringsp.DesDecrypt(rawEncryptMsgBytes, msgkeyBytes)
	return rawMsgBytes, err
}
