package config

import "runtime"

type AppInfo struct {
	AppName string
	Version string
	OS      string
	ARCH    string
	Date    string
}

var appInfo AppInfo

func GetAppInfo() *AppInfo {
	return &appInfo
}

func InitAppInfo(version, date string) {
	appInfo.AppName = "Lucky(大吉)"
	appInfo.Version = version
	appInfo.Date = date
	appInfo.OS = runtime.GOOS
	appInfo.ARCH = runtime.GOARCH
}
