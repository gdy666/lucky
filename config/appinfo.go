package config

import (
	"runtime"
	"time"
)

type AppInfo struct {
	AppName   string
	Version   string
	OS        string
	ARCH      string
	Date      string
	RunTime   string
	GoVersion string
}

var appInfo AppInfo

func GetAppInfo() *AppInfo {
	return &appInfo
}

func InitAppInfo(version, date string) {
	appInfo.AppName = "Lucky"
	appInfo.Version = version
	appInfo.Date = date
	appInfo.OS = runtime.GOOS
	appInfo.ARCH = runtime.GOARCH
	appInfo.RunTime = time.Now().Format("2006-01-02 15:04:05")
	appInfo.GoVersion = runtime.Version()

	time.Now().Format("2006-01-02T15:04:05Z")

	buildTime, err := time.Parse("2006-01-02T15:04:05Z", date)
	if err != nil {
		return
	}
	appInfo.Date = buildTime.Local().Format("2006-01-02 15:04:05")

}
