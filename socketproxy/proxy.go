// Copyright 2022 gdy, 272288813@qq.com
package socketproxy

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/gdy666/lucky/thirdlib/gdylib/pool"
	"github.com/sirupsen/logrus"
)

type Proxy interface {
	StartProxy()
	StopProxy()

	ReceiveDataCallback(int64)
	SendDataCallback(int64)
	GetProxyType() string
	GetStatus() string
	GetListenIP() string
	GetListenPort() int
	GetKey() string
	GetCurrentConnections() int64

	String() string
	GetTrafficIn() int64
	GetTrafficOut() int64
	SafeCheck(ip string) bool
}

type RelayRuleOptions struct {
	UDPPackageSize                                int    `json:"UDPPackageSize,omitempty"`
	SingleProxyMaxTCPConnections                  int64  `json:"SingleProxyMaxTCPConnections,omitempty"`
	SingleProxyMaxUDPReadTargetDatagoroutineCount int64  `json:"SingleProxyMaxUDPReadTargetDatagoroutineCount"`
	UDPProxyPerformanceMode                       bool   `json:"UDPProxyPerformanceMode,omitempty"`
	UDPShortMode                                  bool   `json:"UDPShortMode,omitempty"`
	SafeMode                                      string `json:"SafeMode,omitempty"`
}

// Join two io.ReadWriteCloser and do some operations.
func (p *BaseProxyConf) relayData(targetServer io.ReadWriteCloser, client io.ReadWriteCloser) {
	var wait sync.WaitGroup
	pipe := func(to io.ReadWriteCloser, from io.ReadWriteCloser, writedataCallback func(int64)) {
		defer to.Close()
		defer from.Close()
		defer wait.Done()

		nw, _ := p.copyBuffer(to, from, nil, nil)
		if writedataCallback != nil {
			writedataCallback(nw)
		}

		// if p.TrafficMonitor {
		// 	buf := pool.GetBuf(8 * 1024 * 1024)
		// 	p.CopyBuffer(to, from, buf, writedataCallback)
		// 	pool.PutBuf(buf)
		// } else {
		// 	nw, _ := p.copyBuffer(to, from, nil, nil)
		// 	if writedataCallback != nil {
		// 		writedataCallback(nw)
		// 	}
		// }

	}

	wait.Add(2)
	go pipe(targetServer, client, p.ReceiveDataCallback)
	go pipe(client, targetServer, p.SendDataCallback)
	wait.Wait()
}

func (p *BaseProxyConf) CopyBuffer(dst io.Writer, src io.Reader, buf []byte, writedataCallback func(int64)) (written int64, err error) {
	if buf != nil && len(buf) == 0 {
		panic("empty buffer in CopyBuffer")
	}
	return p.copyBuffer(dst, src, buf, writedataCallback)
}

// copyBuffer is the actual implementation of Copy and CopyBuffer.
// if buf is nil, one is allocated.
func (p *BaseProxyConf) copyBuffer(dst io.Writer, src io.Reader, buf []byte, writedataCallback func(int64)) (written int64, err error) {
	if buf == nil {
		if wt, ok := src.(io.WriterTo); ok {
			return wt.WriteTo(dst)
		}

		if rt, ok := dst.(io.ReaderFrom); ok {
			return rt.ReadFrom(src)
		}

		size := 32 * 1024
		if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
			if l.N < 1 {
				size = 1
			} else {
				size = int(l.N)
			}
		}

		buf = pool.GetBuf(8 * size)
		defer pool.PutBuf(buf)
	}

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			written += int64(nw)

			if writedataCallback != nil {
				writedataCallback(int64(nw))
			}

			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	return written, err
}

func formatFileSize(fileSize int64) (size string) {
	switch {
	case fileSize < 1024:
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	case fileSize < (1024 * 1024):
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	case fileSize < (1024 * 1024 * 1024):
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	case fileSize < (1024 * 1024 * 1024 * 1024):
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	case fileSize < (1024 * 1024 * 1024 * 1024 * 1024):
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	default:
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}

}

func CreateProxy(log *logrus.Logger, proxyType, listenIP string, targetAddressList []string, listenPort, targetPort int, options *RelayRuleOptions) (p Proxy, err error) {
	//key := GetProxyKey(proxyType, listenIP, listenPort)
	switch {
	case strings.HasPrefix(proxyType, "tcp"):
		{
			return CreateTCPProxy(log, proxyType, listenIP, targetAddressList, listenPort, targetPort, options), nil
		}
	case strings.HasPrefix(proxyType, "udp"):
		{
			return CreateUDPProxy(log, proxyType, listenIP, targetAddressList, listenPort, targetPort, options), nil
		}
	default:
		return nil, fmt.Errorf("未支持的类型:%s", proxyType)
	}

}

func GetProxyKey(proxyType, listenIP string, listenPort int) string {
	return fmt.Sprintf("%s@%s:%d", proxyType, listenIP, listenPort)
}
