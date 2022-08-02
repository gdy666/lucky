package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
)

// updateStatusType 更新状态
type updateStatusType string

// ExecWebhook 添加或更新IPv4/IPv6记录
func (d *DDNSTask) ExecWebhook(domains *DomainsState) {
	if !d.WebhookEnable {
		return
	}

	if domains.IpAddr == "" && !d.WebhookCallOnGetIPfail {
		return
	}

	tryUpdate := hasDomainTryToUpdate(domains.Domains)

	if d.WebhookURL != "" && tryUpdate {

		//log.Printf("DDNS任务【%s】触发Webhook", d.TaskName)

		nowTime := time.Now().Format("2006-01-02 15:04:05")

		url := replaceWebhookPara(domains, nowTime, d.WebhookURL)
		requestBody := replaceWebhookPara(domains, nowTime, d.WebhookRequestBody)

		//headersStr := cb.task.DNS.Callback.Headers
		var headerStrList []string
		for i := range d.WebhookHeaders {
			header := replaceWebhookPara(domains, nowTime, d.WebhookHeaders[i])
			headerStrList = append(headerStrList, header)
		}

		headers := httputils.CreateHeadersMap(headerStrList)

		succcssCotentList := []string{}
		for i := range d.WebhookSuccessContent {
			content := replaceWebhookPara(domains, nowTime, d.WebhookSuccessContent[i])
			succcssCotentList = append(succcssCotentList, content)
		}

		callErr := d.webhookHttpClientDo(d.WebhookMethod, url, requestBody, headers, succcssCotentList)

		if callErr != nil {
			//log.Printf("WebHook 调用出错：%s", callErr.Error())
			domains.SetWebhookResult(false, callErr.Error())
			return
		}

		//log.Printf("Webhook 调用成功")
		domains.SetWebhookResult(true, "")

	}
}

func WebhookTest(d *DDNSTask, url, method, WebhookRequestBody, proxy, addr, user, passwd string, headerList, successContentListraw []string) (string, error) {
	nowTime := time.Now().Format("2006-01-02 15:04:05")
	url = replaceWebhookTestPara(url, nowTime)
	requestBody := replaceWebhookTestPara(WebhookRequestBody, nowTime)

	//headersStr := cb.task.DNS.Callback.Headers
	var headerStrList []string
	for i := range headerList {
		header := replaceWebhookTestPara(headerList[i], nowTime)
		headerStrList = append(headerStrList, header)
	}

	headers := httputils.CreateHeadersMap(headerStrList)

	succcssCotentList := []string{}
	for i := range successContentListraw {
		content := replaceWebhookTestPara(successContentListraw[i], nowTime)
		succcssCotentList = append(succcssCotentList, content)
	}

	globalDDNSConf := GetDDNSConfigure()
	proxyType := ""
	proxyAddr := ""
	proxyUser := ""
	proxyPasswd := ""

	switch proxy {
	case "dns":
		{
			if d.DNS.HttpClientProxyType != "" && d.DNS.HttpClientProxyAddr != "" {
				proxyType = d.DNS.HttpClientProxyType
				proxyAddr = d.DNS.HttpClientProxyAddr
				proxyUser = d.DNS.HttpClientProxyUser
				proxyPasswd = d.DNS.HttpClientProxyPassword
			}
		}
	case "http", "https", "socks5":
		{
			proxyType = proxy
			proxyAddr = addr
			proxyUser = user
			proxyPasswd = passwd
		}
	default:
	}

	//fmt.Printf("proxyType:%s\taddr:%s\t,user[%s]passwd[%s]\n", proxyType, proxyAddr, proxyUser, proxyPasswd)

	//dnsConf := cb.task.DNS
	respStr, err := httputils.GetStringGoutDoHttpRequest(
		method,
		url,
		requestBody,
		proxyType,
		proxyAddr,
		proxyUser,
		proxyPasswd,
		headers,
		!globalDDNSConf.HttpClientSecureVerify,
		time.Second*20)
	if err != nil {
		return "", fmt.Errorf("webhookTest 调用接口[%s]出错：%s", url, err.Error())
	}

	for _, successContent := range succcssCotentList {
		if strings.Contains(respStr, successContent) {
			return respStr, nil
		}
	}

	return respStr, fmt.Errorf("接口调用出错,未匹配预设成功返回的字符串")
}

func (d *DDNSTask) webhookHttpClientDo(method, url, requestBody string, headers map[string]string, callbackSuccessContent []string) error {

	globalDDNSConf := GetDDNSConfigure()
	proxyType := ""
	proxyAddr := ""
	proxyUser := ""
	proxyPasswd := ""

	switch d.WebhookProxy {
	case "dns":
		{
			if d.DNS.HttpClientProxyType != "" && d.DNS.HttpClientProxyAddr != "" {
				proxyType = d.DNS.HttpClientProxyType
				proxyAddr = d.DNS.HttpClientProxyAddr
				proxyUser = d.DNS.HttpClientProxyUser
				proxyPasswd = d.DNS.HttpClientProxyPassword
			}
		}
	case "http", "https", "socks5":
		{
			proxyType = d.WebhookProxy
			proxyAddr = d.WebhookProxyAddr
			proxyUser = d.WebhookProxyUser
			proxyPasswd = d.WebhookProxyPassword
		}
	default:
	}

	//dnsConf := cb.task.DNS
	respStr, err := httputils.GetStringGoutDoHttpRequest(
		method,
		url,
		requestBody,
		proxyType,
		proxyAddr,
		proxyUser,
		proxyPasswd,
		headers,
		!globalDDNSConf.HttpClientSecureVerify,
		time.Second*20)
	if err != nil {
		return fmt.Errorf("webhook 调用接口[%s]出错：%s", url, err.Error())
	}

	for _, successContent := range callbackSuccessContent {
		if strings.Contains(respStr, successContent) {

			return nil
		}
	}

	return fmt.Errorf("webhook 调用接口失败:\n%s", respStr)
}

// DomainsIsChange
func hasDomainTryToUpdate(domains []*Domain) bool {
	for _, v46 := range domains {
		switch v46.UpdateStatus {
		case UpdatedFailed:
			return true
		case UpdatedSuccess:
			return true
		default:
		}
	}
	return false
}

// replaceWebhookTestPara WebhookTest替换参数  #{successDomains},#{failedDomains}
func replaceWebhookTestPara(orgPara, nowTime string) (newPara string) {
	orgPara = strings.ReplaceAll(orgPara, "#{ipAddr}", "66.66.66.66")
	orgPara = strings.ReplaceAll(orgPara, "#{successDomains}", "baidu.com,google.com")
	orgPara = strings.ReplaceAll(orgPara, "#{failedDomains}", "weibo.com,github.com")
	orgPara = strings.ReplaceAll(orgPara, "#{time}", nowTime)
	return orgPara
}

// replacePara 替换参数  #{successDomains},#{failedDomains}
func replaceWebhookPara(d *DomainsState, nowTime, orgPara string) (newPara string) {
	ipAddrText := d.IpAddr
	if ipAddrText == "" {
		ipAddrText = "获取IP失败"
	}
	orgPara = strings.ReplaceAll(orgPara, "#{ipAddr}", ipAddrText)
	successDomains, failedDomains := getDomainsStr(d.Domains)
	orgPara = strings.ReplaceAll(orgPara, "#{successDomains}", successDomains)
	orgPara = strings.ReplaceAll(orgPara, "#{failedDomains}", failedDomains)
	orgPara = strings.ReplaceAll(orgPara, "#{time}", nowTime)
	return orgPara
}

// getDomainsStr 用逗号分割域名,分类域名返回，成功和失败的
func getDomainsStr(domains []*Domain) (string, string) {
	var successDomainBuf strings.Builder
	var failedDomainsBuf strings.Builder
	for _, v46 := range domains {
		if v46.UpdateStatus == UpdatedFailed {
			if failedDomainsBuf.Len() > 0 {
				failedDomainsBuf.WriteString(",")
			}
			failedDomainsBuf.WriteString(v46.String())
			continue
		}

		if v46.UpdateStatus == UpdatedNothing || v46.UpdateStatus == UpdatedSuccess {
			if successDomainBuf.Len() > 0 {
				successDomainBuf.WriteString(",")
			}
			successDomainBuf.WriteString(v46.String())
		}
	}

	return successDomainBuf.String(), failedDomainsBuf.String()
}
