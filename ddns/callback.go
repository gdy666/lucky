package ddns

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
)

type Callback struct {
	DNSCommon
	TTL string
}

// Init 初始化
func (cb *Callback) Init(task *config.DDNSTask) {
	cb.DNSCommon.Init(task)

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

func (cb *Callback) createUpdateDomain(recordType, ipAddr string, domain *config.Domain) {

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
		domain.SetDomainUpdateStatus(config.UpdatedFailed, callErr.Error())
		return
	}
	domain.SetDomainUpdateStatus(config.UpdatedSuccess, "")
}

// replacePara 替换参数
func replacePara(orgPara, ipAddr string, domain *config.Domain, recordType string, ttl string) (newPara string) {
	orgPara = strings.ReplaceAll(orgPara, "#{ip}", ipAddr)
	orgPara = strings.ReplaceAll(orgPara, "#{domain}", domain.String())
	orgPara = strings.ReplaceAll(orgPara, "#{recordType}", recordType)
	orgPara = strings.ReplaceAll(orgPara, "#{ttl}", ttl)

	return orgPara
}

func (cb *Callback) CallbackHttpClientDo(method, url, requestBody string, headers map[string]string, callbackSuccessContent []string) error {

	globalDDNSConf := config.GetDDNSConfigure()
	dnsConf := cb.task.DNS
	respStr, err := httputils.GetStringGoutDoHttpRequest(
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
	//log.Printf("接口[%s]调用响应:%s\n", url, respStr)

	for _, successContent := range callbackSuccessContent {
		if strings.Contains(respStr, successContent) {
			return nil
		}
	}

	return fmt.Errorf("调用接口失败:\n%s", respStr)
}
