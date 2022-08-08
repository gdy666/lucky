package ddnscore

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
)

// ExecWebhook 添加或更新IPv4/IPv6记录
func (d *DDNSTaskInfo) ExecWebhook(state *DDNSTaskState) {
	if !d.WebhookEnable {
		return
	}

	if state.IpAddr == "" && !d.WebhookCallOnGetIPfail {
		return
	}

	hasUpdate := hasDomainTryToUpdate(&state.Domains)

	if d.WebhookURL != "" && (hasUpdate || (state.IpAddr == "" && d.WebhookCallOnGetIPfail)) {

		//log.Printf("DDNS任务【%s】触发Webhook", d.TaskName)

		nowTime := time.Now().Format("2006-01-02 15:04:05")

		url := d.replaceWebhookPara(nowTime, d.WebhookURL)
		requestBody := d.replaceWebhookPara(nowTime, d.WebhookRequestBody)

		//headersStr := cb.task.DNS.Callback.Headers
		var headerStrList []string
		for i := range d.WebhookHeaders {
			header := d.replaceWebhookPara(nowTime, d.WebhookHeaders[i])
			headerStrList = append(headerStrList, header)
		}

		headers := httputils.CreateHeadersMap(headerStrList)

		succcssCotentList := []string{}
		for i := range d.WebhookSuccessContent {
			content := d.replaceWebhookPara(nowTime, d.WebhookSuccessContent[i])
			succcssCotentList = append(succcssCotentList, content)
		}

		callErr := d.webhookHttpClientDo(d.WebhookMethod, url, requestBody, headers, succcssCotentList)

		if callErr != nil {
			//log.Printf("WebHook 调用出错：%s", callErr.Error())
			state.SetWebhookResult(false, callErr.Error())
			return
		}

		//log.Printf("Webhook 调用成功")
		state.SetWebhookResult(true, "")

	}
}

func WebhookTest(d *DDNSTaskInfo, url, method, WebhookRequestBody, proxy, addr, user, passwd string, headerList, successContentListraw []string) (string, error) {
	nowTime := time.Now().Format("2006-01-02 15:04:05")
	url = replaceWebhookTestPara(url, nowTime)
	requestBody := replaceWebhookTestPara(WebhookRequestBody, nowTime)

	//log.Printf("requestBody:\n%s", requestBody)

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

	globalDDNSConf := config.GetDDNSConfigure()
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
	_, respStr, err := httputils.GetStringGoutDoHttpRequest(
		"tcp",
		"",
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

func (d *DDNSTaskInfo) webhookHttpClientDo(method, url, requestBody string, headers map[string]string, callbackSuccessContent []string) error {

	globalDDNSConf := config.GetDDNSConfigure()
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
	statusCode, respStr, err := httputils.GetStringGoutDoHttpRequest(
		"tcp",
		"",
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

	if d.WebhookDisableCallbackSuccessContentCheck {
		if statusCode == http.StatusOK {
			return nil
		}
		return fmt.Errorf("webhook调用接口失败:\n statusCode:%d\n%s", statusCode, respStr)
	}

	for _, successContent := range callbackSuccessContent {
		if strings.Contains(respStr, successContent) {

			return nil
		}
	}

	return fmt.Errorf("webhook 调用接口失败:\n%s", respStr)
}

// DomainsIsChange
func hasDomainTryToUpdate(domains *[]Domain) bool {
	for _, v46 := range *domains {
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

	successDomains := "www1.google.com,www2.google.com,www3.google.com,www4.google.com"
	failedDomains := "www1.github.com,www2.github.com,www3.github.com,www4.github.com"
	successDomainsLine := strings.Replace(successDomains, ",", `\n`, -1)
	failedDomainsLine := strings.Replace(failedDomains, ",", `\n`, -1)
	orgPara = strings.ReplaceAll(orgPara, "#{successDomains}", successDomains)
	orgPara = strings.ReplaceAll(orgPara, "#{failedDomains}", failedDomains)
	orgPara = strings.ReplaceAll(orgPara, "#{successDomainsLine}", successDomainsLine)
	orgPara = strings.ReplaceAll(orgPara, "#{failedDomainsLine}", failedDomainsLine)
	orgPara = strings.ReplaceAll(orgPara, "#{time}", nowTime)
	return orgPara
}

// replacePara 替换参数  #{successDomains},#{failedDomains}
func (d *DDNSTaskInfo) replaceWebhookPara(nowTime, orgPara string) (newPara string) {
	ipAddrText := d.TaskState.IpAddr

	successDomains, failedDomains := d.getDomainsStr(&d.TaskState.Domains)
	if ipAddrText == "" {
		ipAddrText = "获取IP失败"
		successDomains = ""
		failedDomains = ""
	}

	successDomainsLine := strings.Replace(successDomains, ",", `\n`, -1)
	failedDomainsLine := strings.Replace(failedDomains, ",", `\n`, -1)

	orgPara = strings.ReplaceAll(orgPara, "#{ipAddr}", ipAddrText)

	orgPara = strings.ReplaceAll(orgPara, "#{successDomains}", successDomains)
	orgPara = strings.ReplaceAll(orgPara, "#{failedDomains}", failedDomains)
	orgPara = strings.ReplaceAll(orgPara, "#{successDomainsLine}", successDomainsLine)
	orgPara = strings.ReplaceAll(orgPara, "#{failedDomainsLine}", failedDomainsLine)
	orgPara = strings.ReplaceAll(orgPara, "#{time}", nowTime)
	return orgPara
}

// getDomainsStr 用逗号分割域名,分类域名返回，成功和失败的
func (d *DDNSTaskInfo) getDomainsStr(domains *[]Domain) (string, string) {
	var successDomainBuf strings.Builder
	var failedDomainsBuf strings.Builder
	for _, v46 := range *domains {
		if v46.UpdateStatus == UpdatedFailed || (d.Webhook.WebhookCallOnGetIPfail && v46.UpdateStatus == UpdatePause) {
			if failedDomainsBuf.Len() > 0 {
				failedDomainsBuf.WriteString(",")
			}
			failedDomainsBuf.WriteString(v46.String())
			continue
		}

		//if v46.UpdateStatus == UpdatedNothing || v46.UpdateStatus == UpdatedSuccess {
		if v46.UpdateStatus == UpdatedSuccess {
			if successDomainBuf.Len() > 0 {
				successDomainBuf.WriteString(",")
			}
			successDomainBuf.WriteString(v46.String())
		}
	}

	return successDomainBuf.String(), failedDomainsBuf.String()
}
