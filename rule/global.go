//Copyright 2022 gdy, 272288813@qq.com
package rule

import (
	"fmt"
	"sync"

	"github.com/ljymc/goports/base"
	"github.com/ljymc/goports/config"
)

var globalRelayRules *[]RelayRule
var globalRelayRulesMutex sync.RWMutex

func SetGlobalRelayRules(rl *[]RelayRule) {
	globalRelayRulesMutex.Lock()
	defer globalRelayRulesMutex.Unlock()
	globalRelayRules = rl
	syncRuleListToConfigure()
}

func AddRuleToGlobalRuleList(sync bool, rl ...RelayRule) (string, error) {
	globalRelayRulesMutex.Lock()
	defer globalRelayRulesMutex.Unlock()

	var err error

	if globalRelayRules == nil {
		var rrl []RelayRule
		globalRelayRules = &rrl
	}

	for i := range rl {
		isExits := false
		for j := range *globalRelayRules {
			if (*globalRelayRules)[j].MainConfigure == rl[i].MainConfigure {
				isExits = true
				if err == nil {
					err = fmt.Errorf("\n\t规则[%s]已存在,不再重复加入规则列表", rl[i].MainConfigure)
				} else {
					err = fmt.Errorf("%s\n\t规则[%s]已存在,不再重复加入规则列表", err.Error(), rl[i].MainConfigure)
				}
				break
			}
		}
		if !isExits {
			*globalRelayRules = append(*globalRelayRules, rl[i])
			if rl[i].From != "cmd" && sync {
				syncRes := ""
				if syncErr := syncRuleListToConfigure(); syncErr != nil {
					syncRes = syncErr.Error()
				}
				return syncRes, nil
			}
		}
	}

	return "", err
}

func AlterRuleInGlobalRuleListByKey(key string, rule *RelayRule) (bool, error) {
	globalRelayRulesMutex.Lock()
	defer globalRelayRulesMutex.Unlock()

	keyIndex := -1

	for i := range *globalRelayRules {
		if (*globalRelayRules)[i].MainConfigure != key {
			continue
		}
		keyIndex = i
		break
	}

	if keyIndex < 0 {
		return true, fmt.Errorf("修改转发规则失败,规则[%s]不存在", key)
	}

	(*globalRelayRules)[keyIndex].Disable()
	(*globalRelayRules)[keyIndex] = *rule

	syncErr := syncRuleListToConfigure()
	syncSuccess := true
	if syncErr != nil {
		syncSuccess = false
	}

	return syncSuccess, nil
}

func DeleteGlobalRuleByKey(key string) (bool, error) {
	globalRelayRulesMutex.Lock()
	defer globalRelayRulesMutex.Unlock()

	if globalRelayRules == nil {
		return true, nil
	}

	deleteIndex := -1

	for i := range *globalRelayRules {
		if (*globalRelayRules)[i].MainConfigure == key {
			deleteIndex = i
			break
		}
	}

	if deleteIndex < 0 {
		return true, fmt.Errorf("relayRule[%s]no found,delete failed", key)
	}
	*globalRelayRules = DeleteRuleSlice(*globalRelayRules, deleteIndex)

	syncError := syncRuleListToConfigure()
	syncSuccess := true
	if syncError != nil {
		syncSuccess = false
	}

	return syncSuccess, nil
}

func GetRuleByMainConfigure(configStr string) *RelayRule {
	globalRelayRulesMutex.RLock()
	defer globalRelayRulesMutex.RUnlock()
	if globalRelayRules == nil {
		return nil
	}
	for i := range *globalRelayRules {
		//fmt.Printf("MainConfigure %s:::%s\n", (*globalRelayRules)[i].MainConfigure, configStr)
		if (*globalRelayRules)[i].MainConfigure == configStr {
			r := (*globalRelayRules)[i]
			return &r
		}
	}
	return nil
}

func GetGlobalEnableProxyCount() int64 {
	globalRelayRulesMutex.RLock()
	defer globalRelayRulesMutex.RUnlock()

	if globalRelayRules == nil {
		return 0
	}
	count := int64(0)
	for _, r := range *globalRelayRules {
		if !r.IsEnable {
			continue
		}
		if r.proxyList == nil {
			continue
		}
		count += int64(len(*r.proxyList))
	}
	return count
}

// func StartAllProxy() {
// 	allProxyListMutex.Lock()
// 	defer allProxyListMutex.Unlock()
// 	if allProxyList == nil {
// 		return
// 	}

// 	for _, p := range *allProxyList {
// 		go p.StartProxy()
// 	}
// }

func EnableAllRelayRule() error {
	globalRelayRulesMutex.RLock()
	defer globalRelayRulesMutex.RUnlock()
	var err error

	if globalRelayRules == nil {
		return nil
	}

	for i := range *globalRelayRules {
		if GetGlobalEnableProxyCount()+(*globalRelayRules)[i].GetProxyCount() <= base.GetGlobalMaxProxyCount() {
			if (*globalRelayRules)[i].From == "cmd" || ((*globalRelayRules)[i].From == "configureFile" && (*globalRelayRules)[i].IsEnable) {
				(*globalRelayRules)[i].Enable()
			}
			continue
		}

		if GetGlobalEnableProxyCount()+(*globalRelayRules)[i].GetProxyCount() > base.DEFAULT_MAX_PROXY_COUNT {
			if err == nil {
				err = fmt.Errorf("\n\t超出代理数最大限制,规则[%s]未启用", (*globalRelayRules)[i].MainConfigure)
			} else {
				err = fmt.Errorf("%s\n\t超出代理数最大限制,规则[%s]未启用", err.Error(), (*globalRelayRules)[i].MainConfigure)
			}
		}
	}
	return err
}

func EnableRelayRuleByKey(key string) (*RelayRule, bool, error) {
	globalRelayRulesMutex.RLock()
	defer globalRelayRulesMutex.RUnlock()

	for i := range *globalRelayRules {
		if (*globalRelayRules)[i].MainConfigure == key {
			(*globalRelayRules)[i].Enable()
			syncErr := syncRuleListToConfigure()
			synsSuccess := true
			if syncErr != nil {
				synsSuccess = false
			}

			return &(*globalRelayRules)[i], synsSuccess, nil
		}
	}

	return nil, true, fmt.Errorf("规则[%s]不存在,开启失败", key)
}

func DisableRelayRuleByKey(key string) (*RelayRule, bool, error) {
	globalRelayRulesMutex.RLock()
	defer globalRelayRulesMutex.RUnlock()

	for i := range *globalRelayRules {
		if (*globalRelayRules)[i].MainConfigure == key {
			(*globalRelayRules)[i].Disable()
			syncErr := syncRuleListToConfigure()
			synsSuccess := true
			if syncErr != nil {
				synsSuccess = false
			}
			return &(*globalRelayRules)[i], synsSuccess, nil
		}
	}

	return nil, true, fmt.Errorf("规则[%s]不存在,停用失败", key)
}

func DeleteRuleSlice(a []RelayRule, deleteIndex int) []RelayRule {
	j := 0
	for i := range a {
		if i != deleteIndex {
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}

//syncRuleListToConfigure 同步规则列表到配置
func syncRuleListToConfigure() error {
	var ruleList []config.ConfigureRelayRule
	for i := range *globalRelayRules {
		if (*globalRelayRules)[i].From == "cmd" {
			continue
		}
		rule := config.ConfigureRelayRule{
			Name:         (*globalRelayRules)[i].Name,
			Configurestr: (*globalRelayRules)[i].MainConfigure,
			Enable:       (*globalRelayRules)[i].IsEnable,
			Options:      (*globalRelayRules)[i].Options}
		ruleList = append(ruleList, rule)
	}
	config.SetConfigRuleList(&ruleList)
	return config.Save()
}
