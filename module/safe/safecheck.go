// Copyright 2022 gdy, 272288813@qq.com
package safe

import (
	"time"

	"github.com/gdy666/lucky/config"
)

func SafeCheck(mode, ip string) bool {
	switch mode {
	case "whitelist":
		//log.Printf("whitelist")
		return whiteListCheck(ip)
	case "blacklist":
		//log.Printf("blacklist")
		return blackListCheck(ip)
	default:
		return false
	}
}

func whiteListCheck(ip string) bool {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	if config.Configure == nil {
		return false
	}

	for _, item := range config.Configure.WhiteListConfigure.WhiteList {

		if !item.Contains(ip) {
			continue
		}

		itemEffectiveTime, err := time.ParseInLocation("2006-01-02 15:04:05", item.EffectiveTime, time.Local)
		if err != nil {
			return false
		}

		if time.Since(itemEffectiveTime) < 0 {
			//log.Printf("CCC")
			return true
		}
		return false
	}

	//log.Printf("DDDD")
	return false
}

func blackListCheck(ip string) bool {
	config.ConfigureMutex.RLock()
	defer config.ConfigureMutex.RUnlock()
	if config.Configure == nil {
		return true
	}

	for _, item := range config.Configure.BlackListConfigure.BlackList {
		if !item.Contains(ip) {
			continue
		}
		itemEffectiveTime, err := time.ParseInLocation("2006-01-02 15:04:05", item.EffectiveTime, time.Local)
		if err != nil {
			return true
		}

		if time.Since(itemEffectiveTime) < 0 {
			return false
		}
		return true
	}

	return true
}
