// Copyright 2022 gdy, 272288813@qq.com
package socketproxy

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
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

func CreateTCPProxy(proxyType, listenIP, targetIP string, balanceTargetAddressList *[]string, listenPort, targetPort int, options *RelayRuleOptions) *TCPProxy {
	p := &TCPProxy{}
	p.ProxyType = proxyType
	p.listenIP = listenIP
	p.listenPort = listenPort
	p.targetIP = targetIP
	p.targetPort = targetPort
	if balanceTargetAddressList != nil {
		p.balanceTargetAddressList = *balanceTargetAddressList
	}

	p.safeMode = options.SafeMode

	p.SetMaxConnections(options.SingleProxyMaxConnections)
	return p
}

func (p *TCPProxy) GetStatus() string {
	return fmt.Sprintf("%s\nactivity connections:[%d]", p.String(), p.GetCurrentConnections())
}

// func (p *TCPProxy) CheckConns() bool {
// 	if GetGlobalTCPConns() >= tcpGlobalMaxConnections || p.GetCurrentConnections() >= p.TcpSingleProxyMaxConns {
// 		// if p.GetGlobalTCPConns() >= tcpGlobalMaxConnections {
// 		// 	log.Println("")
// 		// }
// 		// if p.GetCurrentTCPConns() >= p.TcpSingleProxyMaxConns {
// 		// 	log.Printf("超出单代理TCP限制")
// 		// }
// 		return false
// 	}
// 	return true
// }

func (p *TCPProxy) StartProxy() {
	p.listenConnMutex.Lock()
	defer p.listenConnMutex.Unlock()
	if p.listenConn != nil {
		log.Printf("proxy %s is started", p.String())
		return
	}

	if p.connMap == nil {
		p.connMap = make(map[string]net.Conn)
	}
	ln, err := net.Listen(p.ProxyType, p.GetListentAddress())

	if err != nil {
		if strings.Contains(err.Error(), "Only one usage of each socket address") {
			log.Printf("监听IP端口[%s]已被占用,proxy[%s]启动失败", p.GetListentAddress(), p.String())
		} else {
			log.Printf("Cannot start proxy[%s]:%s", p.String(), err)
		}
		return
	}

	p.listenConn = ln

	log.Printf("[proxy][start][%s]", p.String())

	go func() {
		for {
			newConn, err := ln.Accept()

			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					break
				}
				log.Printf(" Cannot accept connection due to error %s", err.Error())
				continue
			}

			newConnAddr := newConn.RemoteAddr().String()
			if !p.SafeCheck(newConnAddr) {
				log.Printf("[%s]新连接 [%s]安全检查未通过", p.GetKey(), newConnAddr)
				newConn.Close()
				continue
			}
			//log.Printf("[%s]新连接[%s]安全检查通过", p.GetKey(), newConnAddr)

			//fmt.Printf("连接IP:[%s]\n", newConn.RemoteAddr().String())

			//log.Printf("new tdp connection %s@%s [%s]===>%s", p.ProxyType, p.ListentAddress, newConn.RemoteAddr().String(), p.TargetAddress)
			if !p.CheckConnections() {
				//log.Printf("超出最大连接数限制\n")
				p.PrintConnectionsInfo()
				log.Printf("[%s]超出最大连接数限制,不再接受新连接", p.GetKey())
				newConn.Close()
				continue
			}

			p.connMapMutex.Lock()
			p.connMap[newConn.RemoteAddr().String()] = newConn
			p.connMapMutex.Unlock()

			//atomic.AddInt64(&tcpconnections, 1)
			p.AddCurrentConnections(1)
			//fmt.Printf("当前全局TCP连接数:%d\n", p.GetGlobalTCPConns())
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
		log.Printf("[proxy][stop][%s]", p.String())
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
