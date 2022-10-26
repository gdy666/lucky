package reverseproxy

import (
	"fmt"
	"log"

	"github.com/gdy666/lucky/config"
)

func InitReverseProxyServer() {
	ruleList := config.GetReverseProxyRuleList()
	for ruleIndex := range ruleList {
		if ruleList[ruleIndex].Enable {
			startRes := ruleList[ruleIndex].ServerStart()
			if startRes == nil {
				log.Printf("启动反向代理服务[%s]成功", ruleList[ruleIndex].Addr())
			} else {
				log.Printf("启动反向代理服务[%s]失败:%s", ruleList[ruleIndex].Addr(), startRes.Error())
			}
		}
	}
}

func EnableRuleByKey(key string, enable bool) error {

	rule := config.GetReverseProxyRuleByKey(key)

	if rule == nil {
		return fmt.Errorf("GetReverseProxyRuleByKey not found:%s", key)
	}

	if enable {
		err := rule.ServerStart()
		if err != nil {
			log.Printf("启用反向代理规则[%s]出错:%s", rule.Addr(), err.Error())
			config.EnableReverseProxyRuleByKey(key, false)
			return fmt.Errorf("启用反向代理规则[%s]出错:%s", rule.Addr(), err.Error())
		} else {
			log.Printf("启用反向代理规则[%s]成功", rule.Addr())
		}
	} else {
		rule.ServerStop()
		log.Printf("停用反向代理规则[%s]成功", rule.Addr())
	}
	return config.EnableReverseProxyRuleByKey(key, enable)
}

type RuleInfo struct {
	config.ReverseProxyRule
	AccessLogs map[string][]any
}

func GetProxyRuleListInfo() *[]RuleInfo {
	ruleList := config.GetReverseProxyRuleList()
	var res []RuleInfo
	for i := range ruleList {
		//ti := createProxyRuleInfo(nil, ruleList[i])
		var ri RuleInfo
		ri.ReverseProxyRule = *ruleList[i]
		ri.AccessLogs = ruleList[i].GetLastLogs()
		res = append(res, ri)
	}
	return &res
}

func GetAccessLogs(ruleKey, proxyKey string, pageSize, page int) (int, []any) {
	var res []any
	total := 0

	subRule := config.GetSubRuleByKey(ruleKey, proxyKey)
	if subRule == nil {
		return 0, res
	}
	total, res = subRule.GetLogsBuffer().GetLogsByLimit(config.WebLogConvert, pageSize, page)
	return total, res
}
