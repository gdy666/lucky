package service

import (
	"fmt"
	"runtime"
	"time"

	kservice "github.com/kardianos/service"
)

// var globalService Service
var startFunc func()

var listenPort string
var configureFileURL string

func RegisterStartFunc(f func()) {
	startFunc = f
}

func SetListenPort(port int) {
	listenPort = fmt.Sprintf("%d", port)
}

func SetConfigureFile(f string) {
	configureFileURL = f
}

type Service struct {
}

func (p *Service) Start(s kservice.Service) error {
	go p.run()
	return nil
}

func (p *Service) run() {
	go func() {
		<-time.After(time.Second * 2)
		startFunc()
	}()
}

func (p *Service) Stop(s kservice.Service) error {
	return nil
}

func GetServiceState() int {
	serviceStatus := -1
	s, err := GetService()
	if err == nil {
		status, _ := s.Status()
		serviceStatus = int(status)
	}
	return serviceStatus
}

func GetService() (kservice.Service, error) {
	options := make(kservice.KeyValue)
	// if kservice.ChosenSystem().String() == "unix-systemv" {
	// 	options["SysvScript"] = sysvScript
	// }
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("仅支持安装卸载windows服务")
	}

	svcConfig := &kservice.Config{
		Name:        "lucky",
		DisplayName: "lucky",
		Description: "ipv6端口转发,反向代理,DDNS,网络唤醒...",
		Arguments:   []string{"-p", listenPort, "-c", configureFileURL},
		Option:      options,
	}

	prg := &Service{}
	s, err := kservice.New(prg, svcConfig)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// 卸载服务
func UninstallService() error {
	s, err := GetService()
	if err != nil {
		return err
	}

	return s.Uninstall()
}

// 安装服务
func InstallService() error {
	s, err := GetService()

	if err != nil {
		return err
	}

	status, err := s.Status()
	if err != nil && status == kservice.StatusUnknown {
		// 服务未知，创建服务
		if err = s.Install(); err == nil {
			//s.Start()
			//log.Println("安装 lucky 服务成功!")
			return nil
		}
		return fmt.Errorf("安装 lucky 服务失败:%s", err.Error())
	}
	return fmt.Errorf("lucky服务已安装,无需再次安装:下一次系统启动lucky会以服务形式启动.")
}

func Stop() error {
	s, err := GetService()
	if err != nil {
		return err
	}
	return s.Stop()
}

func Start() error {
	s, err := GetService()
	if err != nil {
		return err
	}
	return s.Start()
}

func Restart() error {
	s, err := GetService()
	if err != nil {
		return err
	}
	return s.Restart()
}
