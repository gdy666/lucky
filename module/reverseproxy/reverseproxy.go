package reverseproxy

import (
	"fmt"
	"strings"

	"github.com/gdy666/lucky/config"
	reverseproxyconf "github.com/gdy666/lucky/module/reverseproxy/conf"
	"github.com/gdy666/lucky/module/safe"
	ssl "github.com/gdy666/lucky/module/sslcertficate"
	"github.com/gdy666/lucky/thirdlib/gdylib/logsbuffer"
)

func init() {
	reverseproxyconf.GetValidSSLCertficateList = ssl.GetValidSSLCertficateList
	reverseproxyconf.SafeCheck = safe.SafeCheck
}

// TidyReverseProxyCache 整理反向代理日志缓存
func TidyReverseProxyCache() {
	ruleList := GetReverseProxyRuleList()
	var keyListBuffer strings.Builder
	for _, rule := range ruleList {
		keyListBuffer.WriteString(rule.DefaultProxy.Key)
		keyListBuffer.WriteString(",")
		for _, sr := range rule.ProxyList {
			keyListBuffer.WriteString(sr.Key)
			keyListBuffer.WriteString(",")
		}
	}

	keyListStr := keyListBuffer.String()
	logsbuffer.LogsBufferStoreMu.Lock()
	defer logsbuffer.LogsBufferStoreMu.Unlock()

	var needDeleteKeys []string
	for k := range logsbuffer.LogsBufferStore {
		if !strings.HasPrefix(k, "reverseproxy:") {
			continue
		}

		if len(k) <= 13 {
			continue
		}

		if !strings.Contains(keyListStr, k[13:]) {
			needDeleteKeys = append(needDeleteKeys, k)
		}
	}

	for i := range needDeleteKeys {
		delete(logsbuffer.LogsBufferStore, needDeleteKeys[i])
		reverseproxyconf.ReverseProxyServerStore.Delete(needDeleteKeys[i])
	}

}

//------------------------------------------------------------

func GetReverseProxyRuleList() []*reverseproxyconf.ReverseProxyRule {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()

	var resList []*reverseproxyconf.ReverseProxyRule

	for i := range config.Configure.ReverseProxyRuleList {
		config.Configure.ReverseProxyRuleList[i].Init()
		rule := config.Configure.ReverseProxyRuleList[i]
		resList = append(resList, &rule)
	}
	return resList
}

func GetReverseProxyRuleByKey(ruleKey string) *reverseproxyconf.ReverseProxyRule {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	ruleIndex := -1

	for i := range config.Configure.ReverseProxyRuleList {

		if config.Configure.ReverseProxyRuleList[i].RuleKey == ruleKey {
			ruleIndex = i
			break
		}
	}
	if ruleIndex == -1 {
		return nil
	}

	res := config.Configure.ReverseProxyRuleList[ruleIndex]
	return &res
}

func ReverseProxyRuleListAdd(rule *reverseproxyconf.ReverseProxyRule) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()

	config.Configure.ReverseProxyRuleList = append(config.Configure.ReverseProxyRuleList, *rule)
	return config.Save()
}

func ReverseProxyRuleListDelete(ruleKey string) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()

	ruleIndex := -1

	for i := range config.Configure.ReverseProxyRuleList {
		if config.Configure.ReverseProxyRuleList[i].RuleKey == ruleKey {
			ruleIndex = i
			break
		}
	}

	if ruleIndex == -1 {
		return fmt.Errorf("找不到需要删除的反向代理任务")
	}

	config.Configure.ReverseProxyRuleList = DeleteReverseProxyRuleListlice(config.Configure.ReverseProxyRuleList, ruleIndex)
	return config.Save()
}

func EnableReverseProxyRuleByKey(ruleKey string, enable bool) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	ruleIndex := -1

	for i := range config.Configure.ReverseProxyRuleList {
		if config.Configure.ReverseProxyRuleList[i].RuleKey == ruleKey {
			ruleIndex = i
			break
		}
	}
	if ruleIndex == -1 {
		return fmt.Errorf("开关反向代理规则失败,ruleKey %s 未找到", ruleKey)
	}
	config.Configure.ReverseProxyRuleList[ruleIndex].Enable = enable

	return config.Save()
}

func EnableReverseProxySubRule(ruleKey, proxyKey string, enable bool) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	ruleIndex := -1

	for i := range config.Configure.ReverseProxyRuleList {
		if config.Configure.ReverseProxyRuleList[i].RuleKey == ruleKey {
			ruleIndex = i
			break
		}
	}
	if ruleIndex == -1 {
		return fmt.Errorf("开关反向代理子规则失败,ruleKey %s 未找到", ruleKey)
	}

	proxyIndex := -1
	for i := range config.Configure.ReverseProxyRuleList[ruleIndex].ProxyList {
		if config.Configure.ReverseProxyRuleList[ruleIndex].ProxyList[i].Key == proxyKey {
			proxyIndex = i
			break
		}
	}

	if proxyIndex == -1 {
		return fmt.Errorf("开关反向代理子规则失败,proxyKey %s 未找到", proxyKey)
	}

	config.Configure.ReverseProxyRuleList[ruleIndex].ProxyList[proxyIndex].Enable = enable

	return config.Save()

}

func UpdateReverseProxyRulet(rule reverseproxyconf.ReverseProxyRule) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	ruleIndex := -1

	for i := range config.Configure.ReverseProxyRuleList {
		if config.Configure.ReverseProxyRuleList[i].RuleKey == rule.RuleKey {
			ruleIndex = i
			break
		}
	}

	if ruleIndex == -1 {
		return fmt.Errorf("找不到需要更新的反向代理规则")
	}

	//	rule.RuleKey = programConfigure.ReverseProxyRuleList[ruleIndex].RuleKey
	config.Configure.ReverseProxyRuleList[ruleIndex] = rule

	return config.Save()
}

func DeleteReverseProxyRuleListlice(a []reverseproxyconf.ReverseProxyRule, deleteIndex int) []reverseproxyconf.ReverseProxyRule {
	j := 0
	for i := range a {
		if i != deleteIndex {
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}

func GetSubRuleByKey(ruleKey, proxyKey string) *reverseproxyconf.SubReverProxyRule {
	//rule := getSubRuleByKey()

	rule := GetReverseProxyRuleByKey(ruleKey)
	if rule == nil {
		return nil
	}

	//fmt.Printf("FFF ruleKey:%s proxyKey:%s\n", ruleKey, proxyKey)

	if proxyKey == "default" {

		return &rule.DefaultProxy.SubReverProxyRule
	}

	for i := range rule.ProxyList {
		if rule.ProxyList[i].Key == proxyKey {
			return &rule.ProxyList[i].SubReverProxyRule
		}
	}
	return nil
}
