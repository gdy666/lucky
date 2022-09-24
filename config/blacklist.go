// Copyright 2022 gdy, 272288813@qq.com
package config

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type BlackListItem WhiteListItem

func (w *BlackListItem) Contains(ip string) bool {
	netIP := net.ParseIP(ip)
	if netIP == nil {
		return false
	}
	if w.NetIP != nil {
		return w.NetIP.Equal(netIP)
	}

	if w.Cidr != nil {
		return w.Cidr.Contains(netIP)
	}
	return false
}

type BlackListConfigure struct {
	BlackList []BlackListItem `json:"BlackList"` //黑名单列表
}

func GetBlackList() []BlackListItem {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()

	BlackListFlush(false)

	var resList []BlackListItem
	if programConfigure == nil {
		return resList
	}
	for i := range programConfigure.BlackListConfigure.BlackList {
		resList = append(resList, programConfigure.BlackListConfigure.BlackList[i])
	}
	return resList
}

func BlackListInit() {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()
	var netIP net.IP
	var cidr *net.IPNet

	for i := range programConfigure.BlackListConfigure.BlackList {
		netIP = nil
		cidr = nil
		if strings.Contains(programConfigure.BlackListConfigure.BlackList[i].IP, "/") {
			_, cidr, _ = net.ParseCIDR(programConfigure.BlackListConfigure.BlackList[i].IP)
		} else {
			netIP = net.ParseIP(programConfigure.BlackListConfigure.BlackList[i].IP)
		}
		programConfigure.BlackListConfigure.BlackList[i].Cidr = cidr
		programConfigure.BlackListConfigure.BlackList[i].NetIP = netIP
	}
}

func BlackListAdd(ip string, activelifeDuration int32) (string, error) {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()

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

	for i, ipr := range programConfigure.BlackListConfigure.BlackList {
		if ipr.IP == ip {
			programConfigure.BlackListConfigure.BlackList[i].EffectiveTime = EffectiveTimeStr
			return EffectiveTimeStr, Save()
		}
	}
	item := BlackListItem{IP: ip, EffectiveTime: EffectiveTimeStr, NetIP: netIP, Cidr: cidr}
	programConfigure.BlackListConfigure.BlackList = append(programConfigure.BlackListConfigure.BlackList, item)
	return EffectiveTimeStr, Save()
}

func BlackListDelete(ip string) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()

	removeCount := 0
CONTINUECHECK:
	removeIndex := -1

	for i, ipr := range programConfigure.BlackListConfigure.BlackList {
		if ipr.IP == ip {
			removeIndex = i
			break
		}
	}

	if removeIndex >= 0 {
		removeCount++
		programConfigure.BlackListConfigure.BlackList = DeleteBlackListlice(programConfigure.BlackListConfigure.BlackList, removeIndex)
		goto CONTINUECHECK
	}
	if removeCount == 0 {
		return nil
	}
	return Save()
}

func BlackListFlush(lock bool) error {
	if lock {
		programConfigureMutex.Lock()
		defer programConfigureMutex.Unlock()
	}

	removeCount := 0

CONTINUECHECK:
	removeIndex := -1

	for i, ipr := range programConfigure.BlackListConfigure.BlackList {
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
		programConfigure.BlackListConfigure.BlackList = DeleteBlackListlice(programConfigure.BlackListConfigure.BlackList, removeIndex)
		goto CONTINUECHECK
	}

	if removeCount == 0 {
		return nil
	}
	return Save()
}

func DeleteBlackListlice(a []BlackListItem, deleteIndex int) []BlackListItem {
	j := 0
	for i := range a {
		if i != deleteIndex {
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}
