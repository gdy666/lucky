package web

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gdy666/lucky/rule"
	"github.com/gdy666/lucky/socketproxy"
	"github.com/gin-gonic/gin"
)

func rulelist(c *gin.Context) {
	ruleList, proxyListInfoMap := rule.GetRelayRuleList()
	type ruleItem struct {
		Name                     string                       `json:"Name"`
		MainConfigure            string                       `json:"Mainconfigure"`
		RelayType                string                       `json:"RelayType"`
		ListenIP                 string                       `json:"ListenIP"`
		ListenPorts              string                       `json:"ListenPorts"`
		TargetIP                 string                       `json:"TargetIP"`
		TargetPorts              string                       `json:"TargetPorts"`
		BalanceTargetAddressList []string                     `json:"BalanceTargetAddressList"`
		Options                  socketproxy.RelayRuleOptions `json:"Options"`
		SubRuleList              []rule.SubRelayRule          `json:"SubRuleList"`
		From                     string                       `json:"From"`
		IsEnable                 bool                         `json:"Enable"`
		ProxyList                []rule.RelayRuleProxyInfo    `json:"ProxyList"`
	}

	//proxyListInfoMap[(*ruleList)[i].MainConfigure]
	var data []ruleItem

	for i := range *ruleList {
		item := ruleItem{
			Name:                     (*ruleList)[i].Name,
			MainConfigure:            (*ruleList)[i].MainConfigure,
			RelayType:                (*ruleList)[i].RelayType,
			ListenIP:                 (*ruleList)[i].ListenIP,
			ListenPorts:              (*ruleList)[i].ListenPorts,
			TargetIP:                 (*ruleList)[i].TargetIP,
			TargetPorts:              (*ruleList)[i].TargetPorts,
			Options:                  (*ruleList)[i].Options,
			SubRuleList:              (*ruleList)[i].SubRuleList,
			From:                     (*ruleList)[i].From,
			IsEnable:                 (*ruleList)[i].IsEnable,
			ProxyList:                proxyListInfoMap[(*ruleList)[i].MainConfigure],
			BalanceTargetAddressList: (*ruleList)[i].BalanceTargetAddressList,
		}
		data = append(data, item)
	}
	c.JSON(http.StatusOK, gin.H{"ret": 0, "data": data})

}

func addrule(c *gin.Context) {
	var requestRule rule.RelayRule
	err := c.BindJSON(&requestRule)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("请求解析出错:%s", err.Error())})
		return
	}

	dealRequestRule(&requestRule)

	configureStr := requestRule.CreateMainConfigure()

	r, err := rule.CreateRuleByConfigureAndOptions(requestRule.Name, configureStr, requestRule.Options)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("创建转发规则出错:%s", err.Error())})
		return
	}

	synsRes, err := rule.AddRuleToGlobalRuleList(true, *r)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("添加转发规则出错:%s", err.Error())})
		return
	}

	r, _, err = rule.EnableRelayRuleByKey(r.MainConfigure)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": fmt.Sprintf("启用规则出错:%s", err.Error())})
		return
	}
	log.Printf("添加转发规则[%s][%s]成功", r.Name, r.MainConfigure)

	if synsRes != "" {
		synsRes = "保存配置文件出错,请检查配置文件设置"
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": "添加规则并启用成功", "syncres": synsRes})
}

func alterrule(c *gin.Context) {

	var requestRule rule.RelayRule
	err := c.BindJSON(&requestRule)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("修改请求解析出错:%s", err.Error())})
		return
	}

	dealRequestRule(&requestRule)

	//fmt.Printf("balance:%v\n", requestRule.BalanceTargetAddressList)

	preConfigureStr := requestRule.MainConfigure
	configureStr := requestRule.CreateMainConfigure()
	// configureStr := fmt.Sprintf("%s@%s:%sto%s:%s",
	// 	requestRule.RelayType,
	// 	requestRule.ListenIP, requestRule.ListenPorts,
	// 	requestRule.TargetIP, requestRule.TargetPorts)

	r, err := rule.CreateRuleByConfigureAndOptions(requestRule.Name, configureStr, requestRule.Options)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("修改转发规则[%s]时出错:%s", preConfigureStr, err.Error())})
		return
	}

	syncSuccess, err := rule.AlterRuleInGlobalRuleListByKey(preConfigureStr, r)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("修改转发规则[%s]时出错:%s", preConfigureStr, err.Error())})
		return
	}

	r, _, err = rule.EnableRelayRuleByKey(r.MainConfigure)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": fmt.Sprintf("修改转发规则成功,但启用规则时出错:%s", err.Error())})
		return
	}
	log.Printf("修改转发规则[%s][%s]成功", r.Name, r.MainConfigure)

	synsRes := ""

	if !syncSuccess {
		synsRes = "同步修改规则数据到配置文件出错"
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": "修改转发规则成功", "syncres": synsRes})
}

func deleterule(c *gin.Context) {
	ruleKey := c.Query("rule")

	rule.DisableRelayRuleByKey(ruleKey)

	syncSuccess, err := rule.DeleteGlobalRuleByKey(ruleKey)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("删除转发规则出错:%s", err.Error())})
		return
	}

	syncRes := ""
	if !syncSuccess {
		syncRes = "同步规则信息到配置文件出错"
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": "删除成功", "syncres": syncRes})
}

func dealRequestRule(r *rule.RelayRule) {
	r.ListenPorts = strings.TrimSpace(r.ListenPorts)
	r.TargetPorts = strings.TrimSpace(r.TargetPorts)
	r.ListenIP = strings.TrimSpace(r.ListenIP)
	r.TargetIP = strings.TrimSpace(r.TargetIP)
	r.RelayType = strings.TrimSpace(r.RelayType)
	r.Name = strings.TrimSpace(r.Name)

}

func enablerule(c *gin.Context) {

	enable := c.Query("enable")
	key := c.Query("key")

	var err error
	var r *rule.RelayRule
	var syncSuccess bool

	if enable == "true" {
		r, syncSuccess, err = rule.EnableRelayRuleByKey(key)
	} else {
		r, syncSuccess, err = rule.DisableRelayRuleByKey(key)
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ret": 1, "msg": fmt.Sprintf("开关规则出错:%s", err.Error())})
		return
	}

	log.Printf("[%s] relayRule[%s][%s]", enable, r.Name, r.MainConfigure)
	syncRes := ""
	if !syncSuccess {
		syncRes = "同步规则状态到配置文件出错"
	}

	c.JSON(http.StatusOK, gin.H{"ret": 0, "msg": "", "syncres": syncRes})
}
