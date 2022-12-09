// Copyright 2022 gdy, 272288813@qq.com
package safe

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gdy666/lucky/config"
	safeconf "github.com/gdy666/lucky/module/safe/conf"
)

func GetWhiteListBaseConfigure() safeconf.WhiteListBaseConfigure {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	return config.Configure.WhiteListConfigure.BaseConfigure
}

func SetWhiteListBaseConfigure(activelifeDuration int32, url, account, password string) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()
	config.Configure.WhiteListConfigure.BaseConfigure.URL = url
	config.Configure.WhiteListConfigure.BaseConfigure.ActivelifeDuration = activelifeDuration
	config.Configure.WhiteListConfigure.BaseConfigure.BasicAccount = account
	config.Configure.WhiteListConfigure.BaseConfigure.BasicPassword = password
	return config.Save()
}

func GetWhiteList() []safeconf.WhiteListItem {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()

	WhiteListFlush(false)

	var resList []safeconf.WhiteListItem
	if config.Configure == nil {
		return resList
	}
	for i := range config.Configure.WhiteListConfigure.WhiteList {
		resList = append(resList, config.Configure.WhiteListConfigure.WhiteList[i])
	}
	return resList
}

func WhiteListInit() {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	var netIP net.IP
	var cidr *net.IPNet

	for i := range config.Configure.WhiteListConfigure.WhiteList {
		netIP = nil
		cidr = nil
		if strings.Contains(config.Configure.WhiteListConfigure.WhiteList[i].IP, "/") {
			_, cidr, _ = net.ParseCIDR(config.Configure.WhiteListConfigure.WhiteList[i].IP)
		} else {
			netIP = net.ParseIP(config.Configure.WhiteListConfigure.WhiteList[i].IP)
		}
		config.Configure.WhiteListConfigure.WhiteList[i].Cidr = cidr
		config.Configure.WhiteListConfigure.WhiteList[i].NetIP = netIP
	}
}

func WhiteListAdd(ip string, activelifeDuration int32) (string, error) {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()

	var err error
	var netIP net.IP = nil
	var cidr *net.IPNet = nil
	if strings.Contains(ip, "/") {
		_, cidr, err = net.ParseCIDR(ip)
		if err != nil {
			return "", fmt.Errorf("网段格式有误，转换出错：%s", err.Error())
		}
	} else {
		netIP = net.ParseIP(ip)
		if netIP == nil {
			return "", fmt.Errorf("IP格式有误")
		}
	}

	if activelifeDuration <= 0 {
		activelifeDuration = config.Configure.WhiteListConfigure.BaseConfigure.ActivelifeDuration
	}

	EffectiveTimeStr := time.Now().Add(time.Hour * time.Duration(activelifeDuration)).Format("2006-01-02 15:04:05")

	for i, ipr := range config.Configure.WhiteListConfigure.WhiteList {
		if ipr.IP == ip {
			config.Configure.WhiteListConfigure.WhiteList[i].EffectiveTime = EffectiveTimeStr
			return EffectiveTimeStr, config.Save()
		}
	}
	item := safeconf.WhiteListItem{IP: ip, EffectiveTime: EffectiveTimeStr, NetIP: netIP, Cidr: cidr}
	config.Configure.WhiteListConfigure.WhiteList = append(config.Configure.WhiteListConfigure.WhiteList, item)
	return EffectiveTimeStr, config.Save()
}

func WhiteListDelete(ip string) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()

	removeCount := 0
CONTINUECHECK:
	removeIndex := -1

	for i, ipr := range config.Configure.WhiteListConfigure.WhiteList {
		if ipr.IP == ip {
			removeIndex = i
			break
		}
	}

	if removeIndex >= 0 {
		removeCount++
		config.Configure.WhiteListConfigure.WhiteList = DeleteWhiteListlice(config.Configure.WhiteListConfigure.WhiteList, removeIndex)
		goto CONTINUECHECK
	}
	if removeCount == 0 {
		return nil
	}
	return config.Save()
}

func WhiteListFlush(lock bool) error {
	if lock {
		config.ConfigureMutex.Lock()
		defer config.ConfigureMutex.Unlock()
	}

	removeCount := 0

CONTINUECHECK:
	removeIndex := -1

	for i, ipr := range config.Configure.WhiteListConfigure.WhiteList {
		ipat, err := time.ParseInLocation("2006-01-02 15:04:05", ipr.EffectiveTime, time.Local)
		if err != nil { //有效时间格式有误,当失效处理
			removeIndex = i

			break
		}

		if time.Since(ipat) > 0 {
			removeIndex = i
			break
		}
	}

	if removeIndex >= 0 {
		removeCount++
		config.Configure.WhiteListConfigure.WhiteList = DeleteWhiteListlice(config.Configure.WhiteListConfigure.WhiteList, removeIndex)
		goto CONTINUECHECK
	}

	if removeCount == 0 {
		return nil
	}
	return config.Save()
}

func DeleteWhiteListlice(a []safeconf.WhiteListItem, deleteIndex int) []safeconf.WhiteListItem {
	j := 0
	for i := range a {
		if i != deleteIndex {
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}
