//Copyright 2022 gdy, 272288813@qq.com
package base

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"sync/atomic"
)

const TCP_DEFAULT_STREAM_BUFFERSIZE = 128

const DEFAULT_GLOBAL_MAX_CONNECTIONS = int64(10240)
const TCPUDP_DEFAULT_SINGLE_PROXY_MAX_CONNECTIONS = int64(256)
const DEFAULT_MAX_PROXY_COUNT = int64(128)

var globalMaxConnections = DEFAULT_GLOBAL_MAX_CONNECTIONS

var globalCurrentConnections int64 = 0
var gloMaxProxyCount int64 = DEFAULT_MAX_PROXY_COUNT

var safeCheckFunc func(mode, ip string) bool

func SetSafeCheck(f func(mode, ip string) bool) {
	safeCheckFunc = f
}

func SetGlobalMaxProxyCount(max int64) {
	atomic.StoreInt64(&gloMaxProxyCount, max)
}

func GetGlobalMaxProxyCount() int64 {
	return atomic.LoadInt64(&gloMaxProxyCount)
}

func SetGlobalMaxConnections(max int64) {
	atomic.StoreInt64(&globalMaxConnections, max)
}

func GetGlobalMaxConnections() int64 {
	return atomic.LoadInt64(&globalMaxConnections)
}

func GetSingleProxyMaxConnections(m *int64) int64 {
	if *m <= 0 {
		return TCPUDP_DEFAULT_SINGLE_PROXY_MAX_CONNECTIONS
	}
	return *m

}

func GetGlobalConnections() int64 {
	return atomic.LoadInt64(&globalCurrentConnections)
}

func GloBalCOnnectionsAdd(add int64) int64 {
	return atomic.AddInt64(&globalCurrentConnections, add)
}

type TCPUDPProxyCommonConf struct {
	CurrentConnectionsCount   int64
	SingleProxyMaxConnections int64
	targetBalanceIndex        int64
	BaseProxyConf
	listentAddress string
	listenIP       string
	listenPort     int
	targetIP       string
	targetPort     int
	targetAddress  string

	balanceTargetAddressList []string //均衡负载转发
	targetBalanceIndexMutex  sync.Mutex

	safeMode string
}

func (p *TCPUDPProxyCommonConf) CheckConnections() bool {
	if p.GetCurrentConnections() >= GetGlobalMaxConnections() || p.GetCurrentConnections() >= p.SingleProxyMaxConnections {
		return false
	}
	return true
}

func (p *TCPUDPProxyCommonConf) PrintConnectionsInfo() {
	log.Printf("[%s]当前连接数:[%d],单代理最大连接数限制[%d],全局最大连接数限制[%d]\n", p.GetKey(), p.GetCurrentConnections(), p.SingleProxyMaxConnections, GetGlobalMaxConnections())
}

func (p *TCPUDPProxyCommonConf) SetMaxConnections(max int64) {
	if max <= 0 {
		p.SingleProxyMaxConnections = TCPUDP_DEFAULT_SINGLE_PROXY_MAX_CONNECTIONS
	} else {
		p.SingleProxyMaxConnections = max
	}
}

func (p *TCPUDPProxyCommonConf) AddCurrentConnections(a int64) {
	atomic.AddInt64(&p.CurrentConnectionsCount, a)
	GloBalCOnnectionsAdd(a)
}

func (p *TCPUDPProxyCommonConf) GetCurrentConnections() int64 {
	return atomic.LoadInt64(&p.CurrentConnectionsCount)
}

func (p *TCPProxy) GetCurrentCon() int64 {
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
	if len(p.balanceTargetAddressList) == 0 {
		if p.targetAddress == "" {
			if strings.Contains(p.targetIP, ":") {
				p.targetAddress = fmt.Sprintf("[%s]:%d", p.targetIP, p.targetPort)
			} else {
				p.targetAddress = fmt.Sprintf("%s:%d", p.targetIP, p.targetPort)
			}
		}
		return p.targetAddress
	}

	var address string
	addressListLength := int64(len(p.balanceTargetAddressList))
	p.targetBalanceIndexMutex.Lock()
	address = p.balanceTargetAddressList[p.targetBalanceIndex%addressListLength]
	p.targetBalanceIndex++
	p.targetBalanceIndexMutex.Unlock()

	return address
}

func (p *TCPUDPProxyCommonConf) String() string {
	if len(p.balanceTargetAddressList) == 0 {
		return fmt.Sprintf("%s@%s ===> %s", p.ProxyType, p.GetListentAddress(), p.GetTargetAddress())
	}

	return fmt.Sprintf("%s@%s ===> %v", p.ProxyType, p.GetListentAddress(), p.balanceTargetAddressList)
}

func (p *TCPUDPProxyCommonConf) SafeCheck(remodeAddr string) bool {
	host, _, _ := net.SplitHostPort(remodeAddr)
	return safeCheckFunc(p.safeMode, host)
}
