// Copyright 2022 gdy, 272288813@qq.com
package socketproxy

import (
	"sync/atomic"
)

type BaseProxyConf struct {
	TrafficIn  int64
	TrafficOut int64
	key        string
	ProxyType  string // tcp tcp4 tcp6 udp udp4 udp6

}

func (p *BaseProxyConf) GetProxyType() string {
	return p.ProxyType
}

func (p *BaseProxyConf) GetStatus() string {
	return p.ProxyType
}

func (p *BaseProxyConf) ReceiveDataCallback(nw int64) {
	atomic.AddInt64(&p.TrafficIn, nw)
}

func (p *BaseProxyConf) SendDataCallback(nw int64) {
	atomic.AddInt64(&p.TrafficOut, nw)
}

func (p *BaseProxyConf) GetTrafficIn() int64 {
	return atomic.LoadInt64(&p.TrafficIn)
}

func (p *BaseProxyConf) GetTrafficOut() int64 {
	return atomic.LoadInt64(&p.TrafficOut)
}
