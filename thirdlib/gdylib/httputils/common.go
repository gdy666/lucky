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
	"time"

	"golang.org/x/net/proxy"
)

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

//GetStringFromHttpResponse 从response获取
func GetBytesFromHttpResponse(resp *http.Response) ([]byte, error) {
	if resp.Body == nil {
		return []byte{}, nil
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

//GetStringFromHttpResponse 从response获取
func GetStringFromHttpResponse(resp *http.Response) (string, error) {
	respBytes, err := GetBytesFromHttpResponse(resp)
	if err != nil {
		return "", err
	}
	return string(respBytes), nil
}

func NewTransport(secureSkipVerify bool, proxyType, proxyUrl, user, passwd string) (*http.Transport, error) {
	var transport *http.Transport
	proxyType = strings.ToLower(proxyType)
	switch proxyType {
	case "http", "https":
		{
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

			transport = &http.Transport{

				TLSClientConfig: &tls.Config{InsecureSkipVerify: secureSkipVerify},
				Proxy:           http.ProxyURL(urlProxy),
				Dial: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				IdleConnTimeout:     10 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			}
		}
	case "socket5", "socks5":
		{
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
					return dialer.Dial(network, addr)
				},
				Dial: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				IdleConnTimeout:     10 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			}
		}
	default:
		{
			transport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: secureSkipVerify},
				Dial: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				IdleConnTimeout:     10 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			}
		}

	}

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
