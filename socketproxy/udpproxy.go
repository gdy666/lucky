// Copyright 2022 gdy, 272288813@qq.com
package socketproxy

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatedier/golib/errors"
	"github.com/gdy666/lucky/thirdlib/gdylib/pool"
	"github.com/sirupsen/logrus"
)

const UDP_DEFAULT_PACKAGE_SIZE = 1500

//测试

type UDPProxy struct {
	//BaseProxyConf
	TCPUDPProxyCommonConf

	//	targetAddr *net.UDPAddr
	listenConn      *net.UDPConn
	listenConnMutex sync.Mutex

	relayChs []chan *udpPackge
	replyCh  chan *udpPackge

	udpPackageSize int
	//targetudpConnItemMap      map[string]*udpMapItem
	//targetudpConnItemMapMutex sync.RWMutex
	targetConnectSessions                         sync.Map
	Upm                                           bool //性能模式
	ShortMode                                     bool
	isStop                                        bool
	SingleProxyMaxUDPReadTargetDatagoroutineCount int64
}

type udpPackge struct {
	dataSize   int
	data       *[]byte
	remoteAddr *net.UDPAddr
}

type udpTagetConSession struct {
	targetConn *net.UDPConn
	lastTime   time.Time
}

func CreateUDPProxy(log *logrus.Logger, proxyType, listenIP string, targetAddressList []string, listenPort, targetPort int, options *RelayRuleOptions) *UDPProxy {
	p := &UDPProxy{}
	//p.Key = key
	p.ProxyType = proxyType
	p.listenIP = listenIP
	p.listenPort = listenPort
	p.targetAddressList = targetAddressList
	p.targetPort = targetPort

	p.Upm = options.UDPProxyPerformanceMode
	p.ShortMode = options.UDPShortMode
	p.safeMode = options.SafeMode
	p.log = log
	p.SingleProxyMaxUDPReadTargetDatagoroutineCount = options.SingleProxyMaxUDPReadTargetDatagoroutineCount

	p.SetUDPPacketSize(options.UDPPackageSize)
	return p
}

func (p *UDPProxy) getHandlegoroutineNum() int {
	if p.Upm {
		return runtime.NumCPU()
	}
	return 1
}

func (p *UDPProxy) SetUDPPacketSize(size int) {
	if size <= 0 {
		p.udpPackageSize = UDP_DEFAULT_PACKAGE_SIZE
		return
	}
	if size > 65507 {
		p.udpPackageSize = 65507
		return
	}
	p.udpPackageSize = size
}

func (p *UDPProxy) GetUDPPacketSize() int {

	return p.udpPackageSize
}

func (p *UDPProxy) StartProxy() {
	//p.init()
	p.listenConnMutex.Lock()
	defer p.listenConnMutex.Unlock()
	if p.listenConn != nil {
		return
	}

	bindAddr, err := net.ResolveUDPAddr(p.ProxyType, p.GetListentAddress())

	if err != nil {
		p.log.Errorf("Cannot start proxy[%s]:%s", p.GetKey(), err)
		return
	}

	ln, err := net.ListenUDP(p.ProxyType, bindAddr)
	if err != nil {
		if strings.Contains(err.Error(), " bind: Only one usage of each socket address") {
			p.log.Errorf("监听IP端口[%s]已被占用,proxy[%s]启动失败", p.GetListentAddress(), p.String())
		} else {
			p.log.Errorf("Cannot start proxy[%s]:%s", p.String(), err)
		}
		return
	}

	ln.SetReadBuffer(p.getHandlegoroutineNum() * 4 * 1024 * 1024)
	ln.SetWriteBuffer(p.getHandlegoroutineNum() * 4 * 1024 * 1024)

	p.listenConn = ln

	p.log.Infof("[端口转发][开启][%s]", p.String())

	p.relayChs = make([]chan *udpPackge, p.getHandlegoroutineNum())

	for i := range p.relayChs {
		p.relayChs[i] = make(chan *udpPackge, 1024)
	}

	p.replyCh = make(chan *udpPackge, 1024)
	// if p.targetudpConnItemMap == nil {
	// 	p.targetudpConnItemMap = make(map[string]*udpMapItem)
	// }

	for i := range p.relayChs {
		go p.Forwarder(i, p.relayChs[i])
	}

	go p.replyDataToRemotAddress()

	go p.CheckTargetUDPConnectSessions()

	for i := 0; i < p.getHandlegoroutineNum(); i++ {
		go p.ListenHandler(ln)
	}

}

func (p *UDPProxy) StopProxy() {
	p.listenConnMutex.Lock()
	defer p.listenConnMutex.Unlock()
	defer func() {
		p.targetConnectSessions.Range(func(key any, value any) bool {
			session := value.(*udpTagetConSession)
			session.targetConn.Close()
			p.targetConnectSessions.Delete(key)
			return true
		})
		p.log.Infof("[端口转发][关闭][%s]", p.String())
	}()

	if p.listenConn == nil {
		return
	}
	p.listenConn.Close()
	p.listenConn = nil
	p.isStop = true
	close(p.replyCh)
	for i := range p.relayChs {
		close(p.relayChs[i])
	}
}

// ReadFromTargetOnce one clientAddr only read once,short mode eg: udp dns
func (p *UDPProxy) ReadFromTargetOnce() bool {
	if p.targetPort == 53 || p.ShortMode {
		return true
	}
	return false
}

// func (p *UDPProxy) GetStatus() string {
// 	return fmt.Sprintf("%s  max packet size[%d]", p.String(), p.GetUDPPacketSize())
// }

func (p *UDPProxy) ListenHandler(ln *net.UDPConn) {

	inDatabuf := pool.GetBuf(p.GetUDPPacketSize())
	defer pool.PutBuf(inDatabuf)
	i := uint64(0)
	for {
		if p.listenConn == nil {
			break
		}

		inDatabufSize, remoteAddr, err := ln.ReadFromUDP(inDatabuf)
		if err != nil {
			if strings.Contains(err.Error(), `smaller than the datagram`) {
				p.log.Errorf("[%s] UDP包最大长度设置过小,请重新设置", p.GetKey())
			} else {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					p.log.Errorf(" %s ReadFromUDP error:\n%s \n", p.String(), err.Error())
				}
			}
			continue
		}

		remoteAddrStr := remoteAddr.String()
		if !p.SafeCheck(remoteAddrStr) {
			p.log.Warnf("[%s]新连接 [%s]安全检查未通过", p.GetKey(), remoteAddrStr)
			continue
		}

		_, ok := p.targetConnectSessions.Load(remoteAddrStr)
		if !ok {
			p.log.Infof("[%s]新连接 [%s]安全检查通过", p.GetKey(), remoteAddrStr)
		}

		data := pool.GetBuf(inDatabufSize)
		copy(data, inDatabuf[:inDatabufSize])

		inUdpPack := udpPackge{dataSize: inDatabufSize, data: &data, remoteAddr: remoteAddr}

		p.relayChs[i%uint64(p.getHandlegoroutineNum())] <- &inUdpPack
		i++

	}
}

func (p *UDPProxy) handlerDataFromTargetAddress(raddr *net.UDPAddr, tgConn *net.UDPConn) {
	readBuffer := pool.GetBuf(p.GetUDPPacketSize())
	var session *udpTagetConSession
	sessionKey := raddr.String()

	defer func() {
		pool.PutBuf(readBuffer)
		if p.ReadFromTargetOnce() {
			tgConn.Close()
		} else {
			p.targetConnectSessions.Delete(sessionKey)
		}
		p.AddCurrentConnections(-1)
		p.log.Infof("[%s]目标地址[%s]关闭连接[%s]", p.GetKey(), tgConn.RemoteAddr().String(), tgConn.LocalAddr().String())
	}()

	var targetConn *net.UDPConn

	p.AddCurrentConnections(1)
	for {
		targetConn = nil
		session = nil

		timeout := 1200 * time.Millisecond
		if p.ReadFromTargetOnce() {
			timeout = 300 * time.Millisecond
		}

		if p.ReadFromTargetOnce() {
			targetConn = tgConn
		} else {
			se, ok := p.targetConnectSessions.Load(sessionKey)
			if !ok {
				return
			}
			session = se.(*udpTagetConSession)
			targetConn = session.targetConn
		}

		targetConn.SetReadDeadline(time.Now().Add(timeout))
		n, _, err := targetConn.ReadFromUDP(readBuffer)
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, `i/o timeout`) && !p.ReadFromTargetOnce() {
				continue
			}
			if !strings.Contains(errStr, `use of closed network connection`) {
				p.log.Errorf("[%s]targetConn ReadFromUDP error:%s", p.GetKey(), err.Error())
			}
			return
		}

		data := pool.GetBuf(n)
		copy(data, readBuffer[:n])
		udpMsg := udpPackge{dataSize: n, data: &data, remoteAddr: raddr}

		if err = errors.PanicToError(func() {
			select {
			case p.replyCh <- &udpMsg: //转发数据到远程地址
			default:
			}
		}); err != nil {
			return
		}

		if p.ReadFromTargetOnce() { //一次性
			return
		}

		//非一次性，刷新时间或者退出
		_, ok := p.targetConnectSessions.Load(sessionKey)
		if !ok {
			return
		}
	}
}

func (p *UDPProxy) Forwarder(kk int, replych chan *udpPackge) {

	// read from targetAddr and write clientAddr

	var err error

	// read from readCh
	for udpMsg := range replych {
		err = nil
		se, ok := p.targetConnectSessions.Load(udpMsg.remoteAddr.String())

		if !ok {
			err := p.CheckReadTargetDataGoroutineLimit()
			if err != nil {
				p.log.Warnf("[%s]转发中止：%s", p.GetKey(), err.Error())
				continue
			}
		}

		var session *udpTagetConSession
		if ok {
			session = se.(*udpTagetConSession)
		} else {
			session = &udpTagetConSession{}
		}

		if !ok {
			addr := p.GetTargetAddress()
			tgAddr, err := net.ResolveUDPAddr("udp", addr)
			if err != nil {
				p.log.Errorf("[%s]UDP端口转发目标地址[%s]解析出错:%s", p.GetKey(), addr, err.Error())
				pool.PutBuf(*udpMsg.data)
				continue
			}
			targetConn, err := net.DialUDP("udp", nil, tgAddr)
			if err != nil {
				p.log.Errorf("[%s]UDP端口转发目标地址[%s]连接出错:%s", p.GetKey(), addr, err.Error())
				pool.PutBuf(*udpMsg.data)
				continue
			}
			targetConn.SetWriteBuffer(4 * 1024 * 1024)
			targetConn.SetReadBuffer(4 * 1024 * 1024)

			session.targetConn = targetConn
		}
		session.lastTime = time.Now()

		if !p.ReadFromTargetOnce() { //只存储非一次性
			p.targetConnectSessions.Store(udpMsg.remoteAddr.String(), session)
		}

		p.ReceiveDataCallback(int64(udpMsg.dataSize)) //接收流量记录

		_, err = session.targetConn.Write(*udpMsg.data)
		if err != nil {
			p.log.Errorf("[%s]转发数据到目标端口出错：%s", p.GetKey(), err.Error())
			session.targetConn.Close()
			continue
		}
		pool.PutBuf(*udpMsg.data)

		if !ok {
			go p.handlerDataFromTargetAddress(udpMsg.remoteAddr, session.targetConn)
		}

	}

}

func (p *UDPProxy) replyDataToRemotAddress() {
	for msg := range p.replyCh {
		_, err := p.listenConn.WriteToUDP(*(msg.data), msg.remoteAddr)
		pool.PutBuf(*msg.data)
		if err != nil {
			p.log.Errorf("[%s]转发目标端口数据到远程端口出错：%s", p.GetKey(), err.Error())
			continue
		}
		p.SendDataCallback(int64(msg.dataSize)) //发送流量记录
	}
}

func (p *UDPProxy) CheckReadTargetDataGoroutineLimit() error {
	if GetGlobalUDPPortForwardGroutineCount() >= GetGlobalUDPReadTargetDataMaxgoroutineCountLimit() {
		return fmt.Errorf("超出端口转发全局UDP读取目标地址数据协程数限制[%d]", GetGlobalUDPReadTargetDataMaxgoroutineCountLimit())
	}

	if p.GetCurrentConnections() >= p.SingleProxyMaxUDPReadTargetDatagoroutineCount {
		return fmt.Errorf("超出单端口UDP读取目标地址数据协程数限制[%d]", p.SingleProxyMaxUDPReadTargetDatagoroutineCount)
	}
	return nil
}

func (p *UDPProxy) CheckTargetUDPConnectSessions() {
	for {
		<-time.After(time.Second * 1)
		if p.isStop {
			return
		}
		if p.GetCurrentConnections() <= 0 {
			continue
		}

		p.targetConnectSessions.Range(func(key any, value any) bool {
			session := value.(*udpTagetConSession)
			if time.Since(session.lastTime) >= 30*time.Second {
				session.targetConn.Close()
				p.targetConnectSessions.Delete(key)
			}
			return true
		})

	}
}
