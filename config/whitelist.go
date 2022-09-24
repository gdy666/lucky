// Copyright 2022 gdy, 272288813@qq.com
package config

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type WhiteListConfigure struct {
	BaseConfigure WhiteListBaseConfigure `json:"BaseConfigure"`
	WhiteList     []WhiteListItem        `json:"WhiteList"` //白名单列表
}

type WhiteListItem struct {
	IP            string     `json:"IP"`
	EffectiveTime string     `json:"Effectivetime"` //有效时间
	NetIP         net.IP     `json:"-"`
	Cidr          *net.IPNet `json:"-"`
}

func (w *WhiteListItem) Contains(ip string) bool {
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

type WhiteListBaseConfigure struct {
	URL                string `json:"URL"`
	ActivelifeDuration int32  `json:"ActivelifeDuration"` //有效期限,小时
	BasicAccount       string `json:"BasicAccount"`
	BasicPassword      string `json:"BasicPassword"`
}

func GetWhiteListBaseConfigure() WhiteListBaseConfigure {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()
	return programConfigure.WhiteListConfigure.BaseConfigure
}

func SetWhiteListBaseConfigure(activelifeDuration int32, url, account, password string) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()
	programConfigure.WhiteListConfigure.BaseConfigure.URL = url
	programConfigure.WhiteListConfigure.BaseConfigure.ActivelifeDuration = activelifeDuration
	programConfigure.WhiteListConfigure.BaseConfigure.BasicAccount = account
	programConfigure.WhiteListConfigure.BaseConfigure.BasicPassword = password
	return Save()
}

func GetWhiteList() []WhiteListItem {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()

	WhiteListFlush(false)

	var resList []WhiteListItem
	if programConfigure == nil {
		return resList
	}
	for i := range programConfigure.WhiteListConfigure.WhiteList {
		resList = append(resList, programConfigure.WhiteListConfigure.WhiteList[i])
	}
	return resList
}

func WhiteListInit() {
	programConfigureMutex.RLock()
	defer programConfigureMutex.RUnlock()
	var netIP net.IP
	var cidr *net.IPNet

	for i := range programConfigure.WhiteListConfigure.WhiteList {
		netIP = nil
		cidr = nil
		if strings.Contains(programConfigure.WhiteListConfigure.WhiteList[i].IP, "/") {
			_, cidr, _ = net.ParseCIDR(programConfigure.WhiteListConfigure.WhiteList[i].IP)
		} else {
			netIP = net.ParseIP(programConfigure.WhiteListConfigure.WhiteList[i].IP)
		}
		programConfigure.WhiteListConfigure.WhiteList[i].Cidr = cidr
		programConfigure.WhiteListConfigure.WhiteList[i].NetIP = netIP
	}
}

func WhiteListAdd(ip string, activelifeDuration int32) (string, error) {
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
		activelifeDuration = programConfigure.WhiteListConfigure.BaseConfigure.ActivelifeDuration
	}

	EffectiveTimeStr := time.Now().Add(time.Hour * time.Duration(activelifeDuration)).Format("2006-01-02 15:04:05")

	for i, ipr := range programConfigure.WhiteListConfigure.WhiteList {
		if ipr.IP == ip {
			programConfigure.WhiteListConfigure.WhiteList[i].EffectiveTime = EffectiveTimeStr
			return EffectiveTimeStr, Save()
		}
	}
	item := WhiteListItem{IP: ip, EffectiveTime: EffectiveTimeStr, NetIP: netIP, Cidr: cidr}
	programConfigure.WhiteListConfigure.WhiteList = append(programConfigure.WhiteListConfigure.WhiteList, item)
	return EffectiveTimeStr, Save()
}

func WhiteListDelete(ip string) error {
	programConfigureMutex.Lock()
	defer programConfigureMutex.Unlock()

	removeCount := 0
CONTINUECHECK:
	removeIndex := -1

	for i, ipr := range programConfigure.WhiteListConfigure.WhiteList {
		if ipr.IP == ip {
			removeIndex = i
			break
		}
	}

	if removeIndex >= 0 {
		removeCount++
		programConfigure.WhiteListConfigure.WhiteList = DeleteWhiteListlice(programConfigure.WhiteListConfigure.WhiteList, removeIndex)
		goto CONTINUECHECK
	}
	if removeCount == 0 {
		return nil
	}
	return Save()
}

func WhiteListFlush(lock bool) error {
	if lock {
		programConfigureMutex.Lock()
		defer programConfigureMutex.Unlock()
	}

	removeCount := 0

CONTINUECHECK:
	removeIndex := -1

	for i, ipr := range programConfigure.WhiteListConfigure.WhiteList {
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
		programConfigure.WhiteListConfigure.WhiteList = DeleteWhiteListlice(programConfigure.WhiteListConfigure.WhiteList, removeIndex)
		goto CONTINUECHECK
	}

	if removeCount == 0 {
		return nil
	}
	return Save()
}

func DeleteWhiteListlice(a []WhiteListItem, deleteIndex int) []WhiteListItem {
	j := 0
	for i := range a {
		if i != deleteIndex {
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}
