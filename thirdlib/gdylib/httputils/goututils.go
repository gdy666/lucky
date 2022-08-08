package httputils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/guonaihong/gout"
	"github.com/guonaihong/gout/dataflow"
)

func NewGout(transportNetwork, localAddr string, secureSkipVerify bool, proxyType, proxyUrl, user, passwd string, timeout time.Duration) (*dataflow.Gout, error) {
	httpClient, err := CreateHttpClient(transportNetwork, localAddr, secureSkipVerify, proxyType, proxyUrl, user, passwd, timeout)
	if err != nil {
		return nil, fmt.Errorf("CreateHttpClient error:%s", err.Error())
	}

	return gout.New(httpClient), nil
}

func GetAndParseJSONResponseFromGoutDoHttpRequest(transportNetwork, localAddr, method, url, requestBody, proxyType, proxyUrl, user, passwd string,
	headers map[string]string, secureSkipVerify bool, timeout time.Duration, result interface{}) error {
	_, bytes, err := GetBytesFromGoutDoHttpRequest(
		transportNetwork,
		localAddr,
		method,
		url,
		requestBody,
		proxyType,
		proxyUrl, user, passwd, headers, secureSkipVerify, timeout)
	if err != nil {
		return fmt.Errorf("GetBytesFromHttpResponse err:%s", err.Error())
	}
	if len(bytes) > 0 {
		err = json.Unmarshal(bytes, &result)
		if err != nil {
			return fmt.Errorf("GetAndParseJSONResponseFromHttpResponse 解析JSON结果出错：%s", err.Error())
		}
	}
	return nil
}

func GetStringGoutDoHttpRequest(transportNetwork, localAddr, method, url, requestBody, proxyType, proxyUrl, user, passwd string,
	headers map[string]string, secureSkipVerify bool, timeout time.Duration) (int, string, error) {
	statusCode, bytes, err := GetBytesFromGoutDoHttpRequest(transportNetwork, localAddr, method, url, requestBody, proxyType, proxyUrl, user, passwd, headers, secureSkipVerify, timeout)
	if err != nil {
		return 0, "", err
	}
	return statusCode, string(bytes), nil
}

func GetBytesFromGoutDoHttpRequest(transportNetwork, localAddr, method, url, requestBody, proxyType, proxyUrl, user, passwd string,
	headers map[string]string, secureSkipVerify bool, timeout time.Duration) (int, []byte, error) {
	gout, err := NewGout(
		transportNetwork,
		localAddr,
		secureSkipVerify,
		proxyType,
		proxyUrl,
		user,
		passwd, timeout)
	if err != nil {
		return 0, []byte{}, fmt.Errorf("GoutDoHttpRequest err:%s", err.Error())
	}

	switch strings.ToLower(method) {
	case "get":
		gout.GET(url)
	case "post":
		gout.POST(url)
	case "put":
		gout.PUT(url)
	case "delete":
		gout.DELETE(url)
	default:
		return 0, []byte{}, fmt.Errorf("未支持的Callback请求方法:%s", method)
	}

	basicAuthUserName, BasicAuthUserNameOk := headers["BasicAuthUserName"]
	basicAuthPassword, BasicAuthPasswordOk := headers["BasicAuthPassword"]
	if BasicAuthUserNameOk && BasicAuthPasswordOk {
		gout.SetBasicAuth(basicAuthUserName, basicAuthPassword)
	}
	delete(headers, "BasicAuthUserName")
	delete(headers, "BasicAuthPassword")

	if len(requestBody) > 0 && method != "get" {
		if json.Valid([]byte(requestBody)) {
			gout.SetJSON(requestBody)
		} else {
			gout.SetWWWForm(requestBody)
		}
	}

	gout.SetHeader(headers)
	//gout.SetTimeout(timeout)

	resp, err := gout.Response()
	if err != nil {
		return 0, []byte{}, fmt.Errorf("gout.Response() error:%s", err.Error())
	}
	respByte, respErr := GetBytesFromHttpResponse(resp)

	return resp.StatusCode, respByte, respErr
}
