package httputils

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/proxy"
)

var globalTransportMap map[string]*http.Transport
var globalTransportMapMutex sync.Mutex

func init() {
	globalTransportMap = make(map[string]*http.Transport)
}

func SplitHostPort(hostPort string) (host, port string) {
	host = hostPort

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return
}

func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

func GetAndParseJSONResponseFromHttpResponse(resp *http.Response, result interface{}) error {
	bytes, err := GetBytesFromHttpResponse(resp)
	if err != nil {
		return fmt.Errorf("GetBytesFromHttpResponse err:%s", err.Error())
	}
	if len(bytes) > 0 {
		err = json.Unmarshal(bytes, &result)
		if err != nil {
			//log.Printf("请求接口解析json结果失败! ERROR: %s\n", err)
			return fmt.Errorf("GetAndParseJSONResponseFromHttpResponse 解析JSON结果出错：%s", err.Error())
		}
	}
	return nil
}

// GetStringFromHttpResponse 从response获取
func GetBytesFromHttpResponse(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return []byte{}, fmt.Errorf("resp.Body = nil")
	}
	defer resp.Body.Close()
	var body []byte
	var err error
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return []byte{}, err
		}
		body, err = ioutil.ReadAll(reader)
		return body, err
	}
	body, err = ioutil.ReadAll(resp.Body)
	return body, err
}

// GetStringFromHttpResponse 从response获取
func GetStringFromHttpResponse(resp *http.Response) (string, error) {
	respBytes, err := GetBytesFromHttpResponse(resp)
	if err != nil {
		return "", err
	}
	return string(respBytes), nil
}

type transportIt struct {
	Network          string
	LocalAddr        string
	ProxyType        string
	ProxyUrl         string
	User             string
	Passwd           string
	SecureSkipVerify bool
}

func (t *transportIt) String() string {
	return fmt.Sprintf("%s,%s,%s,%s,%s,%s,%v", t.Network, t.LocalAddr, t.ProxyType, t.ProxyUrl, t.User, t.Passwd, t.SecureSkipVerify)
}

// NewTransport
// transportNetwork 网络类型 tcp tcp4 tcp6
// localAddr 指定网卡出口
func NewTransport(transportNetwork,
	localAddrStr string,
	secureSkipVerify bool,
	proxyType,
	proxyUrl,
	user,
	passwd string) (*http.Transport, error) {
	var transport *http.Transport
	proxyType = strings.ToLower(proxyType)
	ti := transportIt{
		Network:          transportNetwork,
		LocalAddr:        localAddrStr,
		ProxyType:        proxyType,
		ProxyUrl:         proxyUrl,
		User:             user,
		Passwd:           passwd,
		SecureSkipVerify: secureSkipVerify}

	globalTransportMapMutex.Lock()
	defer globalTransportMapMutex.Unlock()

	tr, ok := globalTransportMap[ti.String()]
	if ok {
		//log.Printf("map[%s]已存在", ti.String())
		return tr, nil
	}
	//log.Printf("map[%s]未存在", ti.String())

	switch proxyType {
	case "http", "https":
		{
			//log.Printf("http proxy Transport network:%s", transportNetwork)
			if !strings.Contains(proxyUrl, "http") {
				proxyUrl = fmt.Sprintf("%s://%s", proxyType, proxyUrl)
			}
			urlProxy, err := url.Parse(proxyUrl)
			if err != nil {
				return nil, fmt.Errorf("NewTransport=>proxy url.Parse error:%s", err.Error())
			}

			if user != "" && passwd != "" {
				urlProxy.User = url.UserPassword(user, passwd)
			}

			var localAddr net.Addr
			localAddr = nil

			if localAddrStr != "" {
				lAddr, err := net.ResolveTCPAddr(transportNetwork, localAddrStr+":0")
				if err != nil {
					return nil, fmt.Errorf("NewTransport=> ResolveTCPAddr localAddr:%s error:%s", localAddrStr, err.Error())
				}
				localAddr = lAddr
			}

			dialer := (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				LocalAddr: localAddr,
			})

			transport = &http.Transport{

				TLSClientConfig: &tls.Config{InsecureSkipVerify: secureSkipVerify},
				Proxy:           http.ProxyURL(urlProxy),
				DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
					return dialer.Dial(transportNetwork, addr)
				},
				ForceAttemptHTTP2:     true,
				Dial:                  dialer.Dial,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}
		}
	case "socket5", "socks5":
		{
			//log.Printf("socket5 proxy Transport network:%s", transportNetwork)
			var userAuth proxy.Auth
			if user != "" && passwd != "" {
				userAuth.User = user
				userAuth.Password = passwd
			}

			dialer, err := proxy.SOCKS5("tcp", proxyUrl, &userAuth, proxy.Direct)
			if err != nil {
				return nil, fmt.Errorf("NewTransport=>proxy.SOCKS5 error:%s", err.Error())
			}

			transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: secureSkipVerify},
				DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
					return dialer.Dial(transportNetwork, addr)
				},
				Dial: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}
		}
	default:
		{
			//log.Printf("default Transport network:%s", transportNetwork)
			var localAddr net.Addr

			localAddr = nil
			if localAddrStr != "" {
				lAddr, err := net.ResolveTCPAddr(transportNetwork, localAddrStr+":0")
				if err != nil {
					return nil, fmt.Errorf("NewTransport=> ResolveTCPAddr localAddr:%s error:%s", localAddrStr, err.Error())
				}
				localAddr = lAddr
			}

			dialer := (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				LocalAddr: localAddr,
				Control: func(network, address string, c syscall.RawConn) error {
					//	fmt.Printf("network:%s\taddress:%s\n", network, address)
					if network != transportNetwork && transportNetwork != "tcp" {
						return fmt.Errorf("must use :%s", transportNetwork)
					}
					return nil
				},
			})

			transport = &http.Transport{
				DisableKeepAlives: true,
				TLSClientConfig:   &tls.Config{InsecureSkipVerify: secureSkipVerify},
				DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
					return dialer.Dial(transportNetwork, addr)
				},
				Dial:                  dialer.Dial,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}
		}

	}

	globalTransportMap[ti.String()] = transport

	return transport, nil
}

func CreateHeadersMap(headers []string) map[string]string {
	hm := make(map[string]string)
	for _, header := range headers {
		kvSpliteIndex := strings.Index(header, ":")
		if kvSpliteIndex < 0 {
			continue
		}
		if kvSpliteIndex+1 > len(header) {
			continue
		}
		key := header[:kvSpliteIndex]
		value := header[kvSpliteIndex+1:]
		hm[key] = value
	}

	return hm
}
