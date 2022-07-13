//Copyright 2022 gdy, 272288813@qq.com
package rule

import (
	"github.com/ljymc/goports/base"
)

type RelayRuleProxyInfo struct {
	Proxy              string `json:"Proxy"`
	CurrentConnections int64  `json:"CurrentConnections"`
	TrafficIn          int64  `json:"TrafficIn"`
	TrafficOut         int64  `json:"TrafficOut"`
}

func GetRelayRuleList() (*[]RelayRule, map[string][]RelayRuleProxyInfo) {
	globalRelayRulesMutex.RLock()
	defer globalRelayRulesMutex.RUnlock()
	var rl []RelayRule
	proxyInfoMap := make(map[string][]RelayRuleProxyInfo)
	if globalRelayRules == nil {
		return &rl, proxyInfoMap
	}

	for i := range *globalRelayRules {
		rl = append(rl, (*globalRelayRules)[i])
		var rpl []RelayRuleProxyInfo
		if (*globalRelayRules)[i].proxyList == nil {
			proxyInfoMap[(*globalRelayRules)[i].MainConfigure] = rpl
			continue
		}
		for pindex := range *(*globalRelayRules)[i].proxyList {
			pi := RelayRuleProxyInfo{
				Proxy:              (*(*globalRelayRules)[i].proxyList)[pindex].GetKey(),
				CurrentConnections: (*(*globalRelayRules)[i].proxyList)[pindex].GetCurrentConnections(),
				TrafficIn:          (*(*globalRelayRules)[i].proxyList)[pindex].GetTrafficIn(),
				TrafficOut:         (*(*globalRelayRules)[i].proxyList)[pindex].GetTrafficOut()}
			rpl = append(rpl, pi)
		}
		proxyInfoMap[(*globalRelayRules)[i].MainConfigure] = rpl
	}

	return &rl, proxyInfoMap
}

//GetAllProxyInfo
// func GetAllProxyInfo() map[string]interface{} {
// 	allProxyListMutex.Lock()
// 	defer allProxyListMutex.Unlock()
// 	info := make(map[string]interface{})
// 	if allProxyList == nil {
// 		return info
// 	}
// 	for _, p := range *allProxyList {
// 		pi := GetProxyInfo(p)
// 		info[p.GetKey()] = pi
// 	}

// 	return info
// }

func GetProxyInfo(p base.Proxy) map[string]string {
	pi := make(map[string]string)
	pi["proxyType"] = p.GetProxyType()
	pi["key"] = p.GetKey()
	pi["status"] = p.GetStatus()
	pi["fromRule"] = p.FromRule()
	return pi
}
