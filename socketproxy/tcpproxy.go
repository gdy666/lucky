// Copyright 2022 gdy, 272288813@qq.com
package socketproxy

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

type TCPProxy struct {
	TCPUDPProxyCommonConf
	//TcpSingleProxyMaxConns int64
	//	tcpCurrentConns        int64
	listenConn      net.Listener
	listenConnMutex sync.Mutex

	connMap      map[string]net.Conn
	connMapMutex sync.Mutex
}

func CreateTCPProxy(log *logrus.Logger, proxyType, listenIP string, targetAddressList []string, listenPort, targetPort int, options *RelayRuleOptions) *TCPProxy {
	p := &TCPProxy{}
	p.ProxyType = proxyType
	p.listenIP = listenIP
	p.listenPort = listenPort
	p.targetAddressList = targetAddressList
	p.targetPort = targetPort
	p.log = log

	p.safeMode = options.SafeMode

	p.SetMaxConnections(options.SingleProxyMaxTCPConnections)
	return p
}

func (p *TCPProxy) GetStatus() string {
	return fmt.Sprintf("%s\nactivity connections:[%d]", p.String(), p.GetCurrentConnections())
}

func (p *TCPProxy) CheckConnectionsLimit() error {

	if GetGlobalTCPPortForwardConnections() >= GetGlobalTCPPortforwardMaxConnections() {
		return fmt.Errorf("超出TCP最大总连接数[%d]限制", GetGlobalTCPPortforwardMaxConnections())
	}

	if p.GetCurrentConnections() >= p.SingleProxyMaxConnections {
		return fmt.Errorf("超出单端口TCP最大连接数[%d]限制", p.SingleProxyMaxConnections)
	}

	//全局,单端口限制
	return nil
}

func (p *TCPProxy) StartProxy() {
	p.listenConnMutex.Lock()
	defer p.listenConnMutex.Unlock()
	if p.listenConn != nil {
		//log.Printf("proxy %s is started", p.String())
		p.log.Warnf("proxy %s is started", p.String())
		return
	}

	if p.connMap == nil {
		p.connMap = make(map[string]net.Conn)
	}
	ln, err := net.Listen(p.ProxyType, p.GetListentAddress())

	if err != nil {
		if strings.Contains(err.Error(), "Only one usage of each socket address") {
			p.log.Errorf("监听IP端口[%s]已被占用,proxy[%s]启动失败", p.GetListentAddress(), p.String())
		} else {
			p.log.Errorf("Cannot start proxy[%s]:%s", p.String(), err)
		}
		return
	}

	p.listenConn = ln

	p.log.Infof("[端口转发][开启][%s]", p.String())

	go func() {
		for {
			newConn, err := ln.Accept()

			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					break
				}
				p.log.Errorf(" Cannot accept connection due to error %s", err.Error())
				continue
			}

			err = p.CheckConnectionsLimit()
			if err != nil {
				//p.PrintConnectionsInfo()
				p.log.Warnf("[%s]超出最大连接数限制,不再接受新连接:%s", p.GetKey(), err.Error())
				newConn.Close()
				continue
			}

			newConnAddr := newConn.RemoteAddr().String()
			if !p.SafeCheck(newConnAddr) {
				p.log.Warnf("[%s]新连接 [%s]安全检查未通过", p.GetKey(), newConnAddr)
				newConn.Close()
				continue
			}

			p.log.Infof("[%s]新连接[%s]安全检查通过", p.GetKey(), newConnAddr)

			p.connMapMutex.Lock()
			p.connMap[newConn.RemoteAddr().String()] = newConn
			p.connMapMutex.Unlock()

			p.AddCurrentConnections(1)
			go p.handle(newConn)
		}
	}()

	//
	//p.test()

	//go p.test()
	//go p.tcptest()

}

func (p *TCPProxy) StopProxy() {
	p.listenConnMutex.Lock()
	defer p.listenConnMutex.Unlock()
	defer func() {
		p.log.Infof("[端口转发][关闭][%s]", p.String())
	}()
	if p.listenConn == nil {
		return
	}

	p.listenConn.Close()
	p.listenConn = nil

	p.connMapMutex.Lock()
	for _, conn := range p.connMap {
		conn.Close()
	}
	p.connMap = make(map[string]net.Conn)
	p.connMapMutex.Unlock()
}

func (p *TCPProxy) handle(conn net.Conn) {
	//dialer := net.Dialer{Timeout: 10 * time.Second}
	//targetConn, err := dialer.Dial("tcp", p.TargetAddress)
	targetConn, err := net.Dial("tcp", p.GetTargetAddress())

	defer func() {
		if targetConn != nil {
			targetConn.Close()
		}
		defer conn.Close()
		p.AddCurrentConnections(-1)

		p.connMapMutex.Lock()
		delete(p.connMap, conn.RemoteAddr().String())
		p.log.Infof("[%s]%s 断开连接", p.GetKey(), conn.RemoteAddr().String())
		p.connMapMutex.Unlock()

	}()

	if err != nil {
		log.Printf("%s error:%s", p.String(), err.Error())
		return
	}

	//targetConn.SetDeadline(time.Now().Add(time.Second * 3))

	// targetTcpConn, ok := targetConn.(*net.TCPConn)
	// if ok {
	// 	targetTcpConn.SetReadBuffer(p.BufferSize * 1024 * 256 * 1024)
	// 	targetTcpConn.SetWriteBuffer(p.BufferSize * 1024 * 256 * 1024)
	// }

	p.relayData(targetConn, conn)

}
