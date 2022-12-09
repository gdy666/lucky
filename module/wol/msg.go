package wol

import (
	wolconf "github.com/gdy666/lucky/module/wol/conf"
)

const (
	TypeLogin               = '0'
	TypeLoginResp           = '1'
	TypeSyncClientConfigure = '2'
	TypeReplyWakeUp         = '3'
	TypeShutDown            = '4'
)

var (
	msgTypeMap = map[byte]interface{}{
		TypeLogin:               Login{},
		TypeLoginResp:           LoginResp{},
		TypeSyncClientConfigure: SyncClientConfigure{},
		TypeReplyWakeUp:         ReplyWakeUp{},
		TypeShutDown:            ShutDown{},
	}
)

type Login struct {
	wolconf.WOLClientConfigure
	ClientTimeStamp int64
}

// 服务器发送给客户端,登录认证反馈 ,ret 为0成功,其它失败
type LoginResp struct {
	Ret int //
	Msg string
}

type SyncClientConfigure struct {
	wolconf.WOLClientConfigure
}

type ReplyWakeUp struct {
	MacList      []string
	BroadcastIPs []string
	Port         int
	Repeat       int
}

type ShutDown struct {
}
