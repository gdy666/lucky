package web

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/socketproxy"
	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
	"github.com/gin-gonic/gin"
)

type ruleInfo struct {
	config.PortForwardsRule
	ProxyList []proxyInfo
	LastLogs  []any
}
type proxyInfo struct {
	Proxy              string
	TrafficIn          int64
	TrafficOut         int64
	CurrentConnections int64
}

func PortForwardsRuleList(c *gin.Context) {
	ruleRawList := config.GetPortForwardsRuleList()

	var ruleList []ruleInfo

	for i := range ruleRawList {
		var proxyInfoList []proxyInfo
		for j := range *ruleRawList[i].ReverseProxyList {
			p := proxyInfo{
				Proxy:              (*ruleRawList[i].ReverseProxyList)[j].String(),
				TrafficIn:          (*ruleRawList[i].ReverseProxyList)[j].GetTrafficIn(),
				TrafficOut:         (*ruleRawList[i].ReverseProxyList)[j].GetTrafficOut(),
				CurrentConnections: (*ruleRawList[i].ReverseProxyList)[j].GetCurrentConnections()}
			proxyInfoList = append(proxyInfoList, p)
		}
		r := ruleInfo{
			PortForwardsRule: ruleRawList[i],
			ProxyList:        proxyInfoList,
			LastLogs:         ruleRawList[i].GetLastLogs(ruleRawList[i].WebListShowLastLogMaxCount)}
		ruleList = append(ruleList, r)
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "list": ruleList})
}

func PortForwardsRuleAdd(c *gin.Context) {
	var newRule config.PortForwardsRule
	err := c.Bind(&newRule)
	if err != nil {
		log.Printf("请求解析出错:%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("请求解析出错:%s", err.Error())})
		return
	}

	newRule.Key = stringsp.GetRandomString(16)
	err = newRule.InitProxyList()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": fmt.Sprintf("添加转发规则过程初始化ProxyList出错:%s", err.Error())})
		return
	}

	if int64(config.GetPortForwardsGlobalProxyCount()+newRule.ProxyCount()) > socketproxy.GetGlobalMaxPortForwardsCountLimit() {
		c.JSON(http.StatusOK, gin.H{"ret": 3, "msg": "超出全局端口转发最大数量限制"})
		return
	}

	err = config.PortForwardsRuleListAdd(&newRule)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 4, "msg": fmt.Sprintf("添加转发规则出错:%s", err.Error())})
		return
	}

	config.StartAllSocketProxysByRuleKey(newRule.Key)

	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func PortForwardsRuleAlter(c *gin.Context) {
	var alterRule config.PortForwardsRule
	err := c.Bind(&alterRule)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("请求解析出错:%s", err.Error())})
		return
	}

	err = alterRule.InitProxyList()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": fmt.Sprintf("修改转发规则时初始化代理列表出错:%s", err.Error())})
		return
	}

	if int64(config.GetPortForwardsGlobalProxyCountExcept(alterRule.Key)+alterRule.ProxyCount()) > socketproxy.GetGlobalMaxPortForwardsCountLimit() {
		c.JSON(http.StatusOK, gin.H{"ret": 3, "msg": "超出全局端口转发最大数量限制"})
		return
	}

	config.StopAllSocketProxysByRuleKey(alterRule.Key)

	err = config.UpdatePortForwardsRuleToPortForwardsRuleList(alterRule.Key, &alterRule)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 4, "msg": fmt.Sprintf("修改转发规则出错:%s", err.Error())})
		return
	}

	if alterRule.Enable {
		config.StartAllSocketProxysByRuleKey(alterRule.Key)
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func PortForwardsRuleEnable(c *gin.Context) {
	enableStr := c.Query("enable")
	key := c.Query("key")

	var enable bool = false
	if enableStr == "true" {
		enable = true
	}

	err := config.EnablePortForwardsRuleByKey(key, enable)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("开关转发规则出错:%s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func PortForwardsRuleDelete(c *gin.Context) {
	key := c.Query("key")

	err := config.PortForwardsRuleListDelete(key)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("删除转发规则出错:%s", err.Error())})
		return
	}

	config.TidyPortforwardLogsCache()

	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func portforwardConfigure(c *gin.Context) {
	conf := config.GetPortForwardsConfigure()

	c.JSON(http.StatusOK, gin.H{"ret": 0, "configure": conf})
}

func alterPortForwardConfigure(c *gin.Context) {
	var requestObj config.PortForwardsConfigure
	err := c.BindJSON(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": "请求解析出错"})
		return
	}

	err = config.SetPortForwardsConfigure(&requestObj)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 2, "msg": "保存配置过程发生错误,请检测相关启动配置"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0})
}

func getPortwardRuleLogs(c *gin.Context) {
	key := c.Query("key")
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if pageSize <= 0 {
		pageSize = 10
	}
	page, _ := strconv.Atoi(c.Query("page"))
	if page <= 0 {
		page = 1
	}

	rule := config.GetPortForwardsRuleByKey(key)
	if rule == nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("找不到key:%s对应的规则", key)})
		return
	}
	total, logList := rule.GetLogsBuffer().GetLogsByLimit(config.WebLogConvert, pageSize, page)
	c.JSON(http.StatusOK, gin.H{"ret": 0, "total": total, "page": page, "pageSize": pageSize, "logs": logList})
}
