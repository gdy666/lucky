// Copyright 2022 gdy, 272288813@qq.com
package socketproxy

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

const TCP_DEFAULT_STREAM_BUFFERSIZE = 128
const DEFAULT_GLOBAL_MAX_CONNECTIONS = int64(1024)
const DEFAULT_GLOBAL_UDPReadTargetDataMaxgoroutineCount = int64(1024)
const TCPUDP_DEFAULT_SINGLE_PROXY_MAX_CONNECTIONS = int64(256)

const DEFAULT_MAX_PORTFORWARDS_LIMIT = int64(128)

var globalTCPPortforwardMaxConnectionsLimit = DEFAULT_GLOBAL_MAX_CONNECTIONS
var globalUDPReadTargetDataMaxgoroutineCountLimit = DEFAULT_GLOBAL_UDPReadTargetDataMaxgoroutineCount

var globalTCPPortForwardCurrentConnections int64 = 0
var globalUDPPortForwardCurrentGroutineCount int64 = 0

var gloMaxPortForwardsCountLimit int64 = DEFAULT_MAX_PORTFORWARDS_LIMIT

var safeCheckFunc func(mode, ip string) bool

func SetSafeCheck(f func(mode, ip string) bool) {
	safeCheckFunc = f
}

func SetGlobalUDPReadTargetDataMaxgoroutineCountLimit(max int64) {
	atomic.StoreInt64(&globalUDPReadTargetDataMaxgoroutineCountLimit, max)
}

func GetGlobalUDPReadTargetDataMaxgoroutineCountLimit() int64 {
	return atomic.LoadInt64(&globalUDPReadTargetDataMaxgoroutineCountLimit)
}

func SetGlobalMaxPortForwardsCountLimit(max int64) {
	atomic.StoreInt64(&gloMaxPortForwardsCountLimit, max)
}

func GetGlobalMaxPortForwardsCountLimit() int64 {
	return atomic.LoadInt64(&gloMaxPortForwardsCountLimit)
}

func SetGlobalTCPPortforwardMaxConnections(max int64) {
	atomic.StoreInt64(&globalTCPPortforwardMaxConnectionsLimit, max)
}

func GetGlobalTCPPortforwardMaxConnections() int64 {
	return atomic.LoadInt64(&globalTCPPortforwardMaxConnectionsLimit)
}

func GetGlobalTCPPortForwardConnections() int64 {
	return atomic.LoadInt64(&globalTCPPortForwardCurrentConnections)
}

func GloBalTCPPortForwardConnectionsAdd(add int64) int64 {
	return atomic.AddInt64(&globalTCPPortForwardCurrentConnections, add)
}

func GetGlobalUDPPortForwardGroutineCount() int64 {
	return atomic.LoadInt64(&globalUDPPortForwardCurrentGroutineCount)
}

func GloBalUDPPortForwardGroutineCountAdd(add int64) int64 {
	return atomic.AddInt64(&globalUDPPortForwardCurrentGroutineCount, add)
}

type TCPUDPProxyCommonConf struct {
	CurrentConnectionsCount   int64
	SingleProxyMaxConnections int64

	BaseProxyConf
	listentAddress string
	listenIP       string
	listenPort     int
	//targetIP       string
	targetAddressList  []string
	targetAddressCount int
	targetAddressIndex uint64
	targetAddressLock  sync.Mutex
	targetPort         int

	safeMode string
	log      *logrus.Logger
}

// func (p *TCPUDPProxyCommonConf) PrintConnectionsInfo() {
// 	p.log.Infof("[%s]当前连接数:[%d],当前端口最大TCP连接数限制[%d],全局最大TCP连接数限制[%d]", p.GetKey(), p.GetCurrentConnections(), p.SingleProxyMaxConnections, GetGlobalTCPPortforwardMaxConnections())
// }

func (p *TCPUDPProxyCommonConf) SetMaxConnections(max int64) {
	if max <= 0 {
		p.SingleProxyMaxConnections = TCPUDP_DEFAULT_SINGLE_PROXY_MAX_CONNECTIONS
	} else {
		p.SingleProxyMaxConnections = max
	}
}

func (p *TCPUDPProxyCommonConf) AddCurrentConnections(a int64) {
	atomic.AddInt64(&p.CurrentConnectionsCount, a)
	if strings.HasPrefix(p.ProxyType, "tcp") {
		GloBalTCPPortForwardConnectionsAdd(a)
		return
	}

	if strings.HasPrefix(p.ProxyType, "udp") {
		GloBalUDPPortForwardGroutineCountAdd(a)
		return
	}

}

func (p *TCPUDPProxyCommonConf) GetCurrentConnections() int64 {
	return atomic.LoadInt64(&p.CurrentConnectionsCount)
}

func (p *TCPUDPProxyCommonConf) GetListentAddress() string {
	if p.listentAddress == "" {
		if strings.Contains(p.listenIP, ":") {
			p.listentAddress = fmt.Sprintf("[%s]:%d", p.listenIP, p.listenPort)
		} else {
			p.listentAddress = fmt.Sprintf("%s:%d", p.listenIP, p.listenPort)
		}
	}
	return p.listentAddress
}

func (p *TCPUDPProxyCommonConf) GetKey() string {
	if p.key == "" {
		p.key = GetProxyKey(p.ProxyType, p.listenIP, p.listenPort)
	}
	return p.key
}

func (p *TCPUDPProxyCommonConf) GetListenIP() string {
	return p.listenIP
}

func (p *TCPUDPProxyCommonConf) GetListenPort() int {
	return p.listenPort
}

func (p *TCPUDPProxyCommonConf) GetTargetAddress() string {
	p.targetAddressLock.Lock()
	defer p.targetAddressLock.Unlock()
	if p.targetAddressCount <= 0 {
		p.targetAddressCount = len(p.targetAddressList)
		p.targetAddressIndex = 0
	}
	address := fmt.Sprintf("%s:%d", p.targetAddressList[p.targetAddressIndex%uint64(p.targetAddressCount)], p.targetPort)
	p.targetAddressIndex++
	return address
}

func (p *TCPUDPProxyCommonConf) String() string {
	return fmt.Sprintf("%s@%v ===> %v:%d", p.ProxyType, p.GetListentAddress(), p.targetAddressList, p.targetPort)
}

func (p *TCPUDPProxyCommonConf) SafeCheck(remodeAddr string) bool {
	host, _, _ := net.SplitHostPort(remodeAddr)
	return safeCheckFunc(p.safeMode, host)
}
