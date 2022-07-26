//Copyright 2022 gdy, 272288813@qq.com
package base

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatedier/golib/errors"
	"github.com/gdy666/lucky/thirdlib/gdylib/pool"
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

	udpPackageSize            int
	targetudpConnItemMap      map[string]*udpMapItem
	targetudpConnItemMapMutex sync.RWMutex
	Upm                       bool //性能模式
	ShortMode                 bool
}

type udpPackge struct {
	dataSize   int
	data       *[]byte
	remoteAddr *net.UDPAddr
}

func CreateUDPProxy(proxyType, listenIP, targetIP string, balanceTargetAddressList *[]string, listenPort, targetPort int, options *RelayRuleOptions) *UDPProxy {
	p := &UDPProxy{}
	//p.Key = key
	p.ProxyType = proxyType
	p.listenIP = listenIP
	p.listenPort = listenPort
	p.targetIP = targetIP
	p.targetPort = targetPort
	if balanceTargetAddressList != nil {
		p.balanceTargetAddressList = *balanceTargetAddressList
	}

	p.Upm = options.UDPProxyPerformanceMode
	p.ShortMode = options.UDPShortMode
	p.safeMode = options.SafeMode

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
		log.Printf("Cannot start proxy[%s]:%s", p.GetKey(), err)
		return
	}

	ln, err := net.ListenUDP(p.ProxyType, bindAddr)
	if err != nil {
		if strings.Contains(err.Error(), " bind: Only one usage of each socket address") {
			log.Printf("监听IP端口[%s]已被占用,proxy[%s]启动失败", p.GetListentAddress(), p.String())
		} else {
			log.Printf("Cannot start proxy[%s]:%s", p.String(), err)
		}
		return
	}

	ln.SetReadBuffer(p.getHandlegoroutineNum() * 4 * 1024 * 1024)
	ln.SetWriteBuffer(p.getHandlegoroutineNum() * 4 * 1024 * 1024)

	p.listenConn = ln

	log.Printf("[proxy][start][%s]", p.String())

	// p.targetAddr, err = net.ResolveUDPAddr(p.ProxyType, p.TargetAddress)
	// if err != nil {
	// 	log.Printf("net.ResolveUDPAddr[%s] error:%s", p.TargetAddress, err.Error())
	// 	return
	// }

	//go p.test()

	//p.relayCh = make(chan *udpPackge, 1024)
	p.relayChs = make([]chan *udpPackge, p.getHandlegoroutineNum())

	for i := range p.relayChs {
		p.relayChs[i] = make(chan *udpPackge, 1024)
	}

	p.replyCh = make(chan *udpPackge, 1024)
	if p.targetudpConnItemMap == nil {
		p.targetudpConnItemMap = make(map[string]*udpMapItem)
	}

	for i := range p.relayChs {
		go p.Forwarder(i, p.relayChs[i])
	}

	go p.replyDataToRemotAddress()
	go p.CheckTargetUDPConn()

	for i := 0; i < p.getHandlegoroutineNum(); i++ {
		go p.ListenFunc(ln)
	}

}

func (p *UDPProxy) StopProxy() {
	p.listenConnMutex.Lock()
	defer p.listenConnMutex.Unlock()
	defer func() {
		p.targetudpConnItemMapMutex.Lock()
		for _, v := range p.targetudpConnItemMap {
			v.conn.Close()
		}
		p.targetudpConnItemMap = nil
		p.targetudpConnItemMap = make(map[string]*udpMapItem)
		p.targetudpConnItemMapMutex.Unlock()
		log.Printf("[proxy][stop][%s]", p.String())
	}()

	if p.listenConn == nil {
		return
	}
	p.listenConn.Close()
	p.listenConn = nil

}

//ReadFromTargetOnce one clientAddr only read once,short mode eg: udp dns
func (p *UDPProxy) ReadFromTargetOnce() bool {
	if p.targetPort == 53 || p.ShortMode {
		return true
	}
	return false
}

func (p *UDPProxy) GetStatus() string {
	return fmt.Sprintf("%s  max packet size[%d]", p.String(), p.GetUDPPacketSize())
}

func (p *UDPProxy) ListenFunc(ln *net.UDPConn) {

	inDatabuf := pool.GetBuf(p.GetUDPPacketSize())
	defer pool.PutBuf(inDatabuf)
	i := uint64(0)
	for {
		if p.listenConn == nil {
			break
		}

		inDatabufSize, clientAddr, err := ln.ReadFromUDP(inDatabuf)
		if err != nil {
			if strings.Contains(err.Error(), `smaller than the datagram`) {
				log.Printf("%s ReadFromUDP error,the udp packet size is smaller than the datagram,please use flag '-ups xxx'set  udp packet size \n", p.String())
			} else {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					log.Printf(" %s ReadFromUDP error:\n%s \n", p.String(), err.Error())
				}
			}
			continue
		}

		//fmt.Printf("inDatabufSize:%d\n", inDatabufSize)

		newConnAddr := clientAddr.String()
		if !p.SafeCheck(newConnAddr) {
			log.Printf("[%s]新连接 [%s]安全检查未通过", p.GetKey(), newConnAddr)
			continue
		}

		// var newConOk bool
		// p.targetudpConnItemMapMutex.RLock()
		// _, newConOk = p.targetudpConnItemMap[clientAddr.String()]
		// p.targetudpConnItemMapMutex.RUnlock()
		// if !newConOk {
		// 	log.Printf("new udp connection %s@%s [%s]===>%s", p.ProxyType, p.ListentAddress, clientAddr.String(), p.TargetAddress)
		// }
		//log.Printf("new udp connection %s@%s [%s]===>%s", p.ProxyType, p.ListentAddress, clientAddr.String(), p.TargetAddress)

		data := pool.GetBuf(inDatabufSize)
		copy(data, inDatabuf[:inDatabufSize])

		inUdpPack := udpPackge{dataSize: inDatabufSize, data: &data, remoteAddr: clientAddr}
		//p.relayCh <- &inUdpPack

		p.relayChs[i%uint64(p.getHandlegoroutineNum())] <- &inUdpPack
		i++

	}
}

type udpMapItem struct {
	conn     *net.UDPConn
	lastTime time.Time
}

func (p *UDPProxy) Forwarder(kk int, replych chan *udpPackge) {

	// read from targetAddr and write clientAddr
	readFromtargetAddrFunc := func(raddr *net.UDPAddr, udpItemKey string, tgConn *net.UDPConn) {
		readBuffer := pool.GetBuf(p.GetUDPPacketSize())
		defer func() {
			pool.PutBuf(readBuffer)
			if p.ReadFromTargetOnce() {
				tgConn.Close()
			}
			p.AddCurrentConnections(-1)
		}()

		var targetConn *net.UDPConn
		var udpItem *udpMapItem
		var ok bool
		p.AddCurrentConnections(1)
		for {
			targetConn = nil
			udpItem = nil

			timeout := 1200 * time.Millisecond
			if p.ReadFromTargetOnce() {
				timeout = 30 * time.Millisecond
			}

			if p.ReadFromTargetOnce() {
				targetConn = tgConn
			} else {
				p.targetudpConnItemMapMutex.RLock()
				udpItem, ok = p.targetudpConnItemMap[udpItemKey]
				if !ok {
					p.targetudpConnItemMapMutex.RUnlock()
					return
				}
				p.targetudpConnItemMapMutex.RUnlock()
				targetConn = udpItem.conn
			}

			targetConn.SetReadDeadline(time.Now().Add(timeout))
			n, _, err := targetConn.ReadFromUDP(readBuffer)
			if err != nil {
				errStr := err.Error()

				if strings.Contains(errStr, `i/o timeout`) && !p.ReadFromTargetOnce() {
					continue
				}
				if !strings.Contains(errStr, `use of closed network connection`) {
					log.Printf("targetConn ReadFromUDP error:%s", err.Error())
				}
				return
			}

			data := pool.GetBuf(n)
			copy(data, readBuffer[:n])
			udpMsg := udpPackge{dataSize: n, data: &data, remoteAddr: raddr}

			if err = errors.PanicToError(func() {
				select {
				case p.replyCh <- &udpMsg:
				default:
				}
			}); err != nil {
				return
			}

			if !p.ReadFromTargetOnce() {
				p.targetudpConnItemMapMutex.Lock()
				udpItem, ok := p.targetudpConnItemMap[udpItemKey]
				if ok {
					udpItem.lastTime = time.Now()
				}
				p.targetudpConnItemMapMutex.Unlock()

				if !ok {
					return
				}
			}

			if p.ReadFromTargetOnce() {
				return
			}

		}
	}

	var err error
	var targetConn *net.UDPConn

	// read from readCh
	for udpMsg := range replych {
		err = nil
		targetConn = nil

		//if p.ReadFromTargetOnce()
		p.targetudpConnItemMapMutex.Lock()
		udpConnItem, ok := p.targetudpConnItemMap[udpMsg.remoteAddr.String()]
		if !ok || p.ReadFromTargetOnce() { //??

			tgAddr, err := net.ResolveUDPAddr(p.ProxyType, p.GetTargetAddress())
			if err != nil {
				log.Printf("net.ResolveUDPAddr[%s] error:%s", p.GetTargetAddress(), err.Error())
				p.targetudpConnItemMapMutex.Unlock()
				pool.PutBuf(*udpMsg.data)
				continue
			}

			targetConn, err = net.DialUDP("udp", nil, tgAddr)

			if err != nil {
				p.targetudpConnItemMapMutex.Unlock()
				pool.PutBuf(*udpMsg.data)
				continue
			}
			targetConn.SetWriteBuffer(4 * 1024 * 1024)
			targetConn.SetReadBuffer(4 * 1024 * 1024)

			if !ok && !p.ReadFromTargetOnce() {
				p.AddCurrentConnections(1)
				newItem := udpMapItem{conn: targetConn, lastTime: time.Now()}
				p.targetudpConnItemMap[udpMsg.remoteAddr.String()] = &newItem
				udpConnItem = &newItem
			}

		} else {
			udpConnItem.lastTime = time.Now()
			targetConn = udpConnItem.conn
		}
		p.targetudpConnItemMapMutex.Unlock()

		p.ReceiveDataCallback(int64(udpMsg.dataSize))
		_, err = targetConn.Write(*udpMsg.data)
		if err != nil {
			targetConn.Close()
		}
		pool.PutBuf(*udpMsg.data)

		if !ok || p.ReadFromTargetOnce() {
			go readFromtargetAddrFunc(udpMsg.remoteAddr, udpMsg.remoteAddr.String(), targetConn)
		}

	}

}

func (p *UDPProxy) replyDataToRemotAddress() {
	for msg := range p.replyCh {
		_, err := p.listenConn.WriteToUDP(*(msg.data), msg.remoteAddr)
		pool.PutBuf(*msg.data)
		if err != nil {
			log.Printf("udpConn.WriteToUDP error:%s", err.Error())
			continue
		}
		p.SendDataCallback(int64(msg.dataSize))
	}
}

func (p *UDPProxy) CheckTargetUDPConn() {
	for {
		<-time.After(time.Second * 1)
		// connCout := atomic.LoadInt64(&p.targetudpConnCount)
		// if connCout <= 0 {
		// 	continue
		// }
		if p.GetCurrentConnections() <= 0 {
			continue
		}
		p.targetudpConnItemMapMutex.Lock()

		var deleteList []string

		for k, v := range p.targetudpConnItemMap {
			if time.Since(v.lastTime) >= 30*time.Second {
				v.conn.Close()
				deleteList = append(deleteList, k)
			}
		}

		//fmt.Printf("map:%v\t deleteList:%v\n", p.targetudpConnItemMap, deleteList)

		for i := range deleteList {
			delete(p.targetudpConnItemMap, deleteList[i])
			//log.Printf("删除targetudpConnItemMap [%s]\n", deleteList[i])
			//atomic.AddInt64(&p.targetudpConnCount, -1)
			p.AddCurrentConnections(-1)
		}

		p.targetudpConnItemMapMutex.Unlock()
	}
}
