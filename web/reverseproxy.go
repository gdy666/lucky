package web

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/reverseproxy"
	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
	"github.com/gin-gonic/gin"
)

func reverseProxys(c *gin.Context) {
	proxyRuleList := reverseproxy.GetProxyRuleListInfo()

	c.JSON(http.StatusOK, gin.H{"ret": 0, "list": proxyRuleList})
}

func addReverseProxyRule(c *gin.Context) {
	var requestObj config.ReverseProxyRule
	err := c.BindJSON(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "请求解析出错"})
		return
	}

	err = checkReverseProxyRuleRequest(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": err.Error()})
		return
	}

	err = config.ReverseProxyRuleListAdd(&requestObj)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": fmt.Sprintf("添加反向代理规则失败:%s", err.Error())})
		return
	}

	if requestObj.Enable {
		reverseproxy.EnableRuleByKey(requestObj.RuleKey, true)
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": ""})
}

func alterReverseProxyRule(c *gin.Context) {
	var requestObj config.ReverseProxyRule
	err := c.BindJSON(&requestObj)
	if err != nil {
		fmt.Printf("fff:%s\n", err.Error())
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "请求解析出错"})
		return
	}

	err = checkReverseProxyRuleRequest(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": err.Error()})
		return
	}

	err = config.UpdateReverseProxyRulet(requestObj)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": fmt.Sprintf("修改反向代理规则失败:%s", err.Error())})
		return
	}

	reverseproxy.EnableRuleByKey(requestObj.RuleKey, false)
	//reverseproxy.FlushCache(requestObj.RuleKey)
	if requestObj.Enable {
		reverseproxy.EnableRuleByKey(requestObj.RuleKey, true)
	}

	config.TidyReverseProxyCache()

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": ""})
}

func deleteReverseProxyRule(c *gin.Context) {
	ruleKey := c.Query("key")

	err := reverseproxy.EnableRuleByKey(ruleKey, false)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": fmt.Sprintf("删除反向代理规则出错:%s", err.Error())})
		return
	}

	err = config.ReverseProxyRuleListDelete(ruleKey)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 3, "msg": fmt.Sprintf("删除反向代理规则出错:%s", err.Error())})
		return
	}

	config.TidyReverseProxyCache()

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": ""})
}

func enableReverseProxyRule(c *gin.Context) {

	enableStr := c.Query("enable")
	ruleKey := c.Query("ruleKey")
	proxyKey := c.Query("proxyKey")

	enable := false

	if enableStr == "true" {
		enable = true
	}

	if proxyKey == "" { //开关规则
		err := reverseproxy.EnableRuleByKey(ruleKey, enable)
		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "Only one usage of each socket address") {
				errMsg = "端口已被占用"
			}
			c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": errMsg})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": ""})
		return
	}

	err := config.EnableReverseProxySubRule(ruleKey, proxyKey, enable)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": err.Error()})
		return
	}
	//reverseproxy.FlushCache(ruleKey)
	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": ""})
}

func checkReverseProxyRuleRequest(rule *config.ReverseProxyRule) error {
	// if len(rule.ProxyList) <= 0 {
	// 	return fmt.Errorf("至少添加一条反向代理转发规则")
	// }
	var err error

	if rule.RuleKey == "" {
		rule.RuleKey = stringsp.GetRandomString(16)
	}
	rule.DefaultProxy.Key = rule.RuleKey

	if len(rule.DefaultProxy.Locations) > 0 {
		defaultLocations := []string{}
		for i := range rule.DefaultProxy.Locations {
			scheme, hostname, port, _, e := stringsp.GetHostAndPathFromURL(rule.DefaultProxy.Locations[i])
			if e != nil {
				return fmt.Errorf("默认目标地址[%s]格式有误", rule.DefaultProxy.Locations[i])
			}

			if port != "" {
				port = ":" + port
			}

			defaultLocations = append(defaultLocations, fmt.Sprintf("%s://%s%s", scheme, hostname, port))
		}

		rule.DefaultProxy.Locations = defaultLocations
	}

	if rule.DefaultProxy.AddRemoteIPToHeader && rule.DefaultProxy.AddRemoteIPHeaderKey == "" {
		return fmt.Errorf("追加客户端连接IP到指定Header 启用时,自定义HeaderKey不能为空")
	}

	for i := range rule.ProxyList {
		domainsLength := len(rule.ProxyList[i].Domains)
		if domainsLength <= 0 {
			return fmt.Errorf("第 %d 条反向代理转发规则中域名不能为空", i+1)
		}

		locationsLength := len(rule.ProxyList[i].Locations)
		if locationsLength <= 0 {
			return fmt.Errorf("第 %d 条反向代理转发规则中后端目标地址不能为空", i+1)
		}

		for j := range rule.ProxyList[i].Domains {
			_, hostname, _, _, e := stringsp.GetHostAndPathFromURL(rule.ProxyList[i].Domains[j])
			if e != nil {
				return fmt.Errorf("第 %d 条反向代理转发规则中第 %d 条前端地址/域名[%s]格式有误", i+1, j+1, rule.ProxyList[i].Domains[j])
			}
			rule.ProxyList[i].Domains[j] = hostname
		}

		for j := range rule.ProxyList[i].Locations {
			scheme, hostname, port, _, e := stringsp.GetHostAndPathFromURL(rule.ProxyList[i].Locations[j])
			if e != nil {
				return fmt.Errorf("第 %d 条反向代理转发规则中第 %d 条后端目标地址[%s]格式有误", i+1, j+1, rule.ProxyList[i].Locations[j])
			}
			if port != "" {
				port = ":" + port
			}
			rule.ProxyList[i].Locations[j] = fmt.Sprintf("%s://%s%s", scheme, hostname, port)
		}

		if rule.ProxyList[i].AddRemoteIPToHeader && rule.ProxyList[i].AddRemoteIPHeaderKey == "" {
			return fmt.Errorf("第 %d 条子规则中 追加客户端连接IP到指定Header 启用时,自定义HeaderKey不能为空", i+1)
		}

		if rule.ProxyList[i].Key == "" {
			rule.ProxyList[i].Key = stringsp.GetRandomString(16)
		}

		for j := range rule.ProxyList[i].TrustedCIDRsStrList {
			_, _, err = net.ParseCIDR(rule.ProxyList[i].TrustedCIDRsStrList[j])
			if err != nil {
				return fmt.Errorf("第 %d 条子规则中 TrustedCIDRsStrList[%s]格式有误", i+1, rule.ProxyList[i].TrustedCIDRsStrList[j])
			}
		}

	}

	for i := range rule.DefaultProxy.TrustedCIDRsStrList {
		_, _, err = net.ParseCIDR(rule.DefaultProxy.TrustedCIDRsStrList[i])
		if err != nil {
			return fmt.Errorf("默认子规则中的 TrustedCIDRsStrList[%s]格式有误", rule.DefaultProxy.TrustedCIDRsStrList[i])
		}
	}

	rule.Init()

	return nil
}

func getReverseProxyLog(c *gin.Context) {
	ruleKey := c.Query("ruleKey")
	proxyKey := c.Query("proxyKey")
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if pageSize <= 0 {
		pageSize = 10
	}
	page, _ := strconv.Atoi(c.Query("page"))
	if page <= 0 {
		page = 1
	}

	//last := c.Query("last")

	total, logList := reverseproxy.GetAccessLogs(ruleKey, proxyKey, pageSize, page)

	c.JSON(http.StatusOK, gin.H{"ret": 0, "total": total, "page": page, "pageSize": pageSize, "logs": logList})

}
