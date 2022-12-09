package providers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gdy666/lucky/module/ddns/ddnscore.go"
	"github.com/gdy666/lucky/module/ddns/ddnsgo"
	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
)

type Callback struct {
	ProviderCommon
	TTL string
}

// Init 初始化
func (cb *Callback) Init(task *ddnscore.DDNSTaskInfo) {
	cb.ProviderCommon.Init(task)

	if task.TTL == "" {
		// 默认600
		cb.TTL = "600"
	} else {
		cb.TTL = task.TTL
	}
	cb.SetCreateUpdateDomainFunc(cb.createUpdateDomain)
}

func CopyHeadersMap(sm map[string]string) map[string]string {
	dm := make(map[string]string)

	for k, v := range sm {
		dm[k] = v
	}

	return dm
}

func (cb *Callback) createUpdateDomain(recordType, ipAddr string, domain *ddnscore.Domain) {

	url := replacePara(cb.task.DNS.Callback.URL, ipAddr, domain, recordType, cb.TTL)
	requestBody := replacePara(cb.task.DNS.Callback.RequestBody, ipAddr, domain, recordType, cb.TTL)

	//headersStr := cb.task.DNS.Callback.Headers
	var headerStrList []string
	for i := range cb.task.DNS.Callback.Headers {
		header := replacePara(cb.task.DNS.Callback.Headers[i], ipAddr, domain, recordType, cb.TTL)
		headerStrList = append(headerStrList, header)
	}

	headers := httputils.CreateHeadersMap(headerStrList)

	succcssCotentList := []string{}
	for i := range cb.task.DNS.Callback.CallbackSuccessContent {
		content := replacePara(cb.task.DNS.Callback.CallbackSuccessContent[i], ipAddr, domain, recordType, cb.TTL)
		succcssCotentList = append(succcssCotentList, content)
	}

	callErr := cb.CallbackHttpClientDo(cb.task.DNS.Callback.Method, url, requestBody, headers, succcssCotentList)

	if callErr != nil {
		domain.SetDomainUpdateStatus(ddnscore.UpdatedFailed, callErr.Error())
		return
	}
	domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
}

// replacePara 替换参数
func replacePara(orgPara, ipAddr string, domain *ddnscore.Domain, recordType string, ttl string) (newPara string) {
	orgPara = strings.ReplaceAll(orgPara, "#{ip}", ipAddr)
	orgPara = strings.ReplaceAll(orgPara, "#{domain}", domain.String())
	orgPara = strings.ReplaceAll(orgPara, "#{recordType}", recordType)
	orgPara = strings.ReplaceAll(orgPara, "#{ttl}", ttl)

	for k, v := range domain.GetCustomParams() {
		if len(v) == 1 {
			orgPara = strings.ReplaceAll(orgPara, "#{"+k+"}", v[0])
		}
	}

	return orgPara
}

func (cb *Callback) CallbackHttpClientDo(method, url, requestBody string, headers map[string]string, callbackSuccessContent []string) error {

	globalDDNSConf := ddnsgo.GetDDNSConfigure()
	dnsConf := cb.task.DNS
	statusCode, respStr, err := httputils.GetStringGoutDoHttpRequest(
		"tcp",
		"",
		method,
		url,
		requestBody,
		dnsConf.HttpClientProxyType,
		dnsConf.HttpClientProxyAddr,
		dnsConf.HttpClientProxyUser,
		dnsConf.HttpClientProxyPassword,
		headers,
		!globalDDNSConf.HttpClientSecureVerify,
		time.Duration(cb.task.HttpClientTimeout)*time.Second)
	if err != nil {
		return fmt.Errorf("Callback 调用接口[%s]出错：%s", url, err.Error())
	}

	if cb.task.DNS.Callback.DisableCallbackSuccessContentCheck {
		if statusCode == http.StatusOK {
			return nil
		}
		return fmt.Errorf("调用接口失败:\n statusCode:%d\n%s", statusCode, respStr)
	}

	//log.Printf("接口[%s]调用响应:%s\n", url, respStr)

	//fmt.Printf("statusCode:%d\n", statusCode)

	for _, successContent := range callbackSuccessContent {
		if strings.Contains(respStr, successContent) {
			return nil
		}
	}

	return fmt.Errorf("调用接口失败:\n%s", respStr)
}
