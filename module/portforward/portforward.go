package portforward

import (
	"fmt"
	"log"
	"strings"

	"github.com/gdy666/lucky/config"
	portforwardconf "github.com/gdy666/lucky/module/portforward/conf"
	"github.com/gdy666/lucky/module/portforward/socketproxy"
	"github.com/gdy666/lucky/thirdlib/gdylib/logsbuffer"
)

func Init() {
	PortForwardsRuleListInit()
}

// TidyReverseProxyCache 整理端口转发日志缓存
func TidyPortforwardLogsCache() {
	ruleList := GetPortForwardsRuleList()
	var keyListBuffer strings.Builder
	for _, rule := range ruleList {
		keyListBuffer.WriteString(rule.Key)
		keyListBuffer.WriteString(",")
	}

	keyListStr := keyListBuffer.String()
	logsbuffer.LogsBufferStoreMu.Lock()
	defer logsbuffer.LogsBufferStoreMu.Unlock()

	var needDeleteKeys []string

	for k := range logsbuffer.LogsBufferStore {
		if !strings.HasPrefix(k, "portforward:") {
			continue
		}

		if len(k) <= 13 {
			continue
		}

		if !strings.Contains(keyListStr, k[12:]) {
			needDeleteKeys = append(needDeleteKeys, k)
		}
	}

	for i := range needDeleteKeys {
		delete(logsbuffer.LogsBufferStore, needDeleteKeys[i])
	}

}

func PortForwardsRuleListInit() {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	var err error
	for i := range config.Configure.PortForwardsRuleList {
		err = config.Configure.PortForwardsRuleList[i].InitProxyList()
		if err != nil {
			log.Printf("InitProxyList error:%s\n", err.Error())
		}
		if config.Configure.PortForwardsRuleList[i].Enable {
			config.Configure.PortForwardsRuleList[i].StartAllProxys()
		}
	}
}

func GetPortForwardsRuleList() []portforwardconf.PortForwardsRule {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()

	var resList []portforwardconf.PortForwardsRule

	for i := range config.Configure.PortForwardsRuleList {
		r := config.Configure.PortForwardsRuleList[i]
		resList = append(resList, r)
	}
	return resList
}

func GetPortForwardsGlobalProxyCount() int {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	count := 0
	for i := range config.Configure.PortForwardsRuleList {
		count += config.Configure.PortForwardsRuleList[i].ProxyCount()
	}
	return count
}

func GetPortForwardsGlobalProxyCountExcept(key string) int {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	count := 0
	for i := range config.Configure.PortForwardsRuleList {
		if key == config.Configure.PortForwardsRuleList[i].Key {
			continue
		}
		count += config.Configure.PortForwardsRuleList[i].ProxyCount()
	}
	return count
}

func GetPortForwardsRuleByKey(key string) *portforwardconf.PortForwardsRule {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	index := -1

	for i := range config.Configure.PortForwardsRuleList {
		if config.Configure.PortForwardsRuleList[i].Key == key {
			index = i
			break
		}
	}
	if index == -1 {
		return nil
	}
	res := config.Configure.PortForwardsRuleList[index]
	return &res
}

func StopAllSocketProxysByRuleKey(key string) error {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	index := -1

	for i := range config.Configure.PortForwardsRuleList {
		if config.Configure.PortForwardsRuleList[i].Key == key {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("找不到key:%s对应的规则", key)
	}
	config.Configure.PortForwardsRuleList[index].StopAllProxys()
	return nil
}

func StartAllSocketProxysByRuleKey(key string) error {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	index := -1

	for i := range config.Configure.PortForwardsRuleList {
		if config.Configure.PortForwardsRuleList[i].Key == key {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("找不到key:%s对应的规则", key)
	}
	config.Configure.PortForwardsRuleList[index].StartAllProxys()
	return nil
}

func PortForwardsRuleListAdd(r *portforwardconf.PortForwardsRule) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	r.Enable = true
	config.Configure.PortForwardsRuleList = append(config.Configure.PortForwardsRuleList, *r)
	return config.Save()
}

func PortForwardsRuleListDelete(key string) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()

	index := -1

	for i := range config.Configure.PortForwardsRuleList {
		if config.Configure.PortForwardsRuleList[i].Key == key {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("找不到需要删除的端口转发规则")
	}

	config.Configure.PortForwardsRuleList[index].StopAllProxys()

	config.Configure.PortForwardsRuleList = DeletePortForwardsRuleListSlice(config.Configure.PortForwardsRuleList, index)
	return config.Save()
}

func EnablePortForwardsRuleByKey(key string, enable bool) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	index := -1

	for i := range config.Configure.DDNSTaskList {
		if config.Configure.PortForwardsRuleList[i].Key == key {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("开关端口转发规则失败,key查找失败")
	}

	if enable {
		config.Configure.PortForwardsRuleList[index].StartAllProxys()
	} else {
		config.Configure.PortForwardsRuleList[index].StopAllProxys()
	}

	config.Configure.PortForwardsRuleList[index].Enable = enable
	return config.Save()
}

func UpdatePortForwardsRuleToPortForwardsRuleList(key string, r *portforwardconf.PortForwardsRule) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	index := -1

	for i := range config.Configure.PortForwardsRuleList {
		if config.Configure.PortForwardsRuleList[i].Key == key {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("找不到需要更新的端口转发规则")
	}

	config.Configure.PortForwardsRuleList[index] = *r
	return config.Save()
}

func DeletePortForwardsRuleListSlice(a []portforwardconf.PortForwardsRule, deleteIndex int) []portforwardconf.PortForwardsRule {
	j := 0
	for i := range a {
		if i != deleteIndex {
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}

func GetPortForwardsConfigure() portforwardconf.PortForwardsConfigure {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	conf := config.Configure.PortForwardsConfigure
	return conf
}

func SetPortForwardsConfigure(conf *portforwardconf.PortForwardsConfigure) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()

	if conf.PortForwardsLimit < 0 {
		conf.PortForwardsLimit = 0
	} else if conf.PortForwardsLimit > 1024 {
		conf.PortForwardsLimit = 1024
	}

	if conf.TCPPortforwardMaxConnections < 0 {
		conf.TCPPortforwardMaxConnections = 0
	} else if conf.TCPPortforwardMaxConnections > 4096 {
		conf.TCPPortforwardMaxConnections = 4096
	}

	if conf.UDPReadTargetDataMaxgoroutineCount < 0 {
		conf.UDPReadTargetDataMaxgoroutineCount = 0
	} else if conf.UDPReadTargetDataMaxgoroutineCount > 4096 {
		conf.UDPReadTargetDataMaxgoroutineCount = 4096
	}

	config.Configure.PortForwardsConfigure = *conf

	socketproxy.SetGlobalMaxPortForwardsCountLimit(conf.PortForwardsLimit)
	socketproxy.SetGlobalTCPPortforwardMaxConnections(conf.TCPPortforwardMaxConnections)
	socketproxy.SetGlobalUDPReadTargetDataMaxgoroutineCountLimit(conf.UDPReadTargetDataMaxgoroutineCount)
	return config.Save()
}
