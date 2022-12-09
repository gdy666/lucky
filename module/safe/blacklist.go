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

func GetBlackList() []safeconf.BlackListItem {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()

	BlackListFlush(false)

	var resList []safeconf.BlackListItem
	if config.Configure == nil {
		return resList
	}
	for i := range config.Configure.BlackListConfigure.BlackList {
		resList = append(resList, config.Configure.BlackListConfigure.BlackList[i])
	}
	return resList
}

func BlackListInit() {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	var netIP net.IP
	var cidr *net.IPNet

	for i := range config.Configure.BlackListConfigure.BlackList {
		netIP = nil
		cidr = nil
		if strings.Contains(config.Configure.BlackListConfigure.BlackList[i].IP, "/") {
			_, cidr, _ = net.ParseCIDR(config.Configure.BlackListConfigure.BlackList[i].IP)
		} else {
			netIP = net.ParseIP(config.Configure.BlackListConfigure.BlackList[i].IP)
		}
		config.Configure.BlackListConfigure.BlackList[i].Cidr = cidr
		config.Configure.BlackListConfigure.BlackList[i].NetIP = netIP
	}
}

func BlackListAdd(ip string, activelifeDuration int32) (string, error) {
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
		activelifeDuration = 666666
	}

	EffectiveTimeStr := time.Now().Add(time.Hour * time.Duration(activelifeDuration)).Format("2006-01-02 15:04:05")

	for i, ipr := range config.Configure.BlackListConfigure.BlackList {
		if ipr.IP == ip {
			config.Configure.BlackListConfigure.BlackList[i].EffectiveTime = EffectiveTimeStr
			return EffectiveTimeStr, config.Save()
		}
	}
	item := safeconf.BlackListItem{IP: ip, EffectiveTime: EffectiveTimeStr, NetIP: netIP, Cidr: cidr}
	config.Configure.BlackListConfigure.BlackList = append(config.Configure.BlackListConfigure.BlackList, item)
	return EffectiveTimeStr, config.Save()
}

func BlackListDelete(ip string) error {
	config.ConfigureMutex.Lock()
	defer config.ConfigureMutex.Unlock()

	removeCount := 0
CONTINUECHECK:
	removeIndex := -1

	for i, ipr := range config.Configure.BlackListConfigure.BlackList {
		if ipr.IP == ip {
			removeIndex = i
			break
		}
	}

	if removeIndex >= 0 {
		removeCount++
		config.Configure.BlackListConfigure.BlackList = DeleteBlackListlice(config.Configure.BlackListConfigure.BlackList, removeIndex)
		goto CONTINUECHECK
	}
	if removeCount == 0 {
		return nil
	}
	return config.Save()
}

func BlackListFlush(lock bool) error {
	if lock {
		config.ConfigureMutex.Lock()
		defer config.ConfigureMutex.Unlock()
	}

	removeCount := 0

CONTINUECHECK:
	removeIndex := -1

	for i, ipr := range config.Configure.BlackListConfigure.BlackList {
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
		config.Configure.BlackListConfigure.BlackList = DeleteBlackListlice(config.Configure.BlackListConfigure.BlackList, removeIndex)
		goto CONTINUECHECK
	}

	if removeCount == 0 {
		return nil
	}
	return config.Save()
}

func DeleteBlackListlice(a []safeconf.BlackListItem, deleteIndex int) []safeconf.BlackListItem {
	j := 0
	for i := range a {
		if i != deleteIndex {
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}
