//Copyright 2022 gdy, 272288813@qq.com
package base

import (
	"sync/atomic"
)

type BaseProxyConf struct {
	TrafficIn  int64
	TrafficOut int64
	key        string
	ProxyType  string // tcp tcp4 tcp6 udp udp4 udp6

	//TrafficMonitor bool //流量监控
	fromRule string
}

func (p *BaseProxyConf) GetProxyType() string {
	return p.ProxyType
}

func (p *BaseProxyConf) GetStatus() string {
	return p.ProxyType
}

func (p *BaseProxyConf) SetFromRule(rule string) {
	p.fromRule = rule
}

func (p *BaseProxyConf) FromRule() string {
	return p.fromRule
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
