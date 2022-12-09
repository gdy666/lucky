//Copyright 2022 gdy, 272288813@qq.com

package main

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/module/ddns"
	"github.com/gdy666/lucky/module/ddns/ddnsgo"
	"github.com/gdy666/lucky/module/portforward"
	"github.com/gdy666/lucky/module/portforward/socketproxy"
	"github.com/gdy666/lucky/module/reverseproxy"
	"github.com/gdy666/lucky/module/safe"
	"github.com/gdy666/lucky/module/service"
	ssl "github.com/gdy666/lucky/module/sslcertficate"
	"github.com/gdy666/lucky/module/wol"
	"github.com/gdy666/lucky/web"
	kservice "github.com/kardianos/service"
)

var (
	listenPort       = flag.Int("p", 16601, "http Admin Web listen port ")
	configureFileURL = flag.String("c", "", "configure file url")
	disableService   = flag.Bool("ds", false, "disable service mode ")
)

var (
	runMode = "prod"
	version = "dev"
	commit  = "none"
	date    = "2022-07-27T17:54:45Z"
)

var runTime time.Time

func init() {
	var cstZone = time.FixedZone("CST", 8*3600) // 东八
	time.Local = cstZone

	service.RegisterStartFunc(run)
}

func main() {
	flag.Parse()
	service.SetListenPort(*listenPort)
	service.SetConfigureFile(*configureFileURL)

	s, _ := service.GetService()

	if s != nil && !*disableService {
		status, _ := s.Status()
		//fmt.Printf("status:%d\n", status)
		if status != kservice.StatusUnknown {
			log.Printf("以服务形式运行\n")
			if status == kservice.StatusStopped {
				log.Printf("调用启动lucky windows服务")
				service.Start()
				log.Printf("本窗口5秒后退出,lucky将以windows后台服务方式启动.")
				<-time.After(time.Second * 5)
				os.Exit(0)
			}
			s.Run()
			os.Exit(0)
		}
	}

	run()
	var w sync.WaitGroup
	w.Add(1)
	w.Wait()

	// err := service.UninstallService()
	// if err != nil {
	// 	fmt.Printf("%s\n", err.Error())
	// }
	//

}

func run() {

	config.InitAppInfo(version, date)

	err := config.Read(*configureFileURL)
	if err != nil {
		log.Printf("%s", err.Error())
		log.Printf("载入默认配置以及命令行设定的参数")
		config.LoadDefault(*listenPort)
		if len(*configureFileURL) > 0 {
			err = config.Save()
			if err != nil {
				log.Printf("保存配置到%s出错:%s", *configureFileURL, err.Error())
			}
		}
	}

	gcf := config.GetConfig()

	safe.Init()
	ssl.Init()

	wol.Init(web.GetLogger())

	socketproxy.SetSafeCheck(safe.SafeCheck)

	config.SetRunMode(runMode)
	config.SetVersion(version)
	log.Printf("RunMode:%s\n", runMode)
	log.Printf("version:%s\tcommit %s, built at %s\n", version, commit, date)

	RunAdminWeb(&gcf.BaseConfigure)

	runTime = time.Now()

	portforward.Init()

	ddnsgo.DDNSTaskListConfigureCheck()
	ddnsConf := ddnsgo.GetDDNSConfigure()
	if ddnsConf.Enable {
		go ddns.Run(time.Duration(ddnsConf.FirstCheckDelay)*time.Second, time.Duration(ddnsConf.Intervals)*time.Second)
	}

	reverseproxy.InitReverseProxyServer()

	//main goroutine wait
	// sigs := make(chan os.Signal, 1)
	// exit := make(chan bool, 1)
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// go func() {
	// 	<-sigs
	// 	exit <- true
	// }()
	// <-exit

}
