//Copyright 2022 gdy, 272288813@qq.com

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gdy666/lucky/base"
	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/ddns"
	"github.com/gdy666/lucky/rule"
)

var (
	listenPort       = flag.Int("p", 16601, "http Admin Web listen port ")
	pcl              = flag.Int64("pcl", -1, "global proxy count limit")
	gpmc             = flag.Int64("gpmc", -1, "global  proxy max connections,default(1024)")
	udpPackageSize   = flag.Int("ups", base.UDP_DEFAULT_PACKAGE_SIZE, "udp package max size")
	smc              = flag.Int64("smc", base.TCPUDP_DEFAULT_SINGLE_PROXY_MAX_CONNECTIONS, "signle  proxy max connections,default(128)")
	upm              = flag.Bool("upm", true, "udp proxy Performance Mode open")
	udpshort         = flag.Bool("udpshort", false, "udp short mode,eg dns")
	configureFileURL = flag.String("c", "", "configure file url")
)

var (
	runMode = "prod"
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var runTime time.Time

func main() {
	flag.Parse()
	config.InitAppInfo(version, date)

	err := config.Read(*configureFileURL)
	if err != nil {
		log.Printf("%s", err.Error())
		log.Printf("载入默认配置以及命令行设定的参数")
		config.LoadDefault(*pcl, *listenPort, *gpmc)
		if len(*configureFileURL) > 0 {
			err = config.Save()
			if err != nil {
				log.Printf("保存配置到%s出错:%s", *configureFileURL, err.Error())
			}
		}
	}

	gcf := config.GetConfig()

	//fmt.Printf("*gcf:%v\n", *gcf)

	base.SetSafeCheck(config.SafeCheck)
	base.SetGlobalMaxConnections(gcf.BaseConfigure.GlobalMaxConnections)
	base.SetGlobalMaxProxyCount(gcf.BaseConfigure.ProxyCountLimit)
	config.SetRunMode(runMode)
	config.SetVersion(version)
	log.Printf("RunMode:%s\n", runMode)
	log.Printf("version:%s\tcommit %s, built at %s\n", version, commit, date)
	RunAdminWeb(gcf.BaseConfigure.AdminWebListenPort)

	runTime = time.Now()

	// if *upm {
	// 	log.Printf("udp proxy Performance Mode open ")
	// }

	//log.Printf("Gobal  proxy max connections:[%d] single  proxy max connections:[%d]\n", base.GetGlobalMaxConnections(), base.GetSingleProxyMaxConnections(smc))

	if len(flag.Args()) > 0 {
		LoadRuleListFromCMD(flag.Args())
	}

	LoadRuleFromConfigFile(gcf)

	rule.EnableAllRelayRule() //开启规则

	config.DDNSTaskListTaskDetailsInit()
	ddnsConf := config.GetDDNSConfigure()
	if ddnsConf.Enable {
		ddns.Run(time.Duration(ddnsConf.FirstCheckDelay)*time.Second, time.Duration(ddnsConf.Intervals)*time.Second)
	}

	//ddns.RunTimer(time.Second, time.Second*30)

	//initProxyList()

	//*****************
	// time.Sleep(time.Microsecond * 50)
	// cruuentPath, _ := fileutils.GetCurrentDirectory()

	// panicFile := fmt.Sprintf("%s/relayport_panic.log", cruuentPath)
	// fileutils.PanicRedirect(panicFile)
	//*****************

	//main goroutine wait
	sigs := make(chan os.Signal, 1)
	exit := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		exit <- true
	}()
	<-exit
}

func LoadRuleListFromCMD(args []string) {
	options := base.RelayRuleOptions{UDPPackageSize: *udpPackageSize,
		SingleProxyMaxConnections: *smc,
		UDPProxyPerformanceMode:   *upm,
		UDPShortMode:              *udpshort}

	relayRules, err := rule.GetRelayRulesFromCMD(flag.Args(), &options)
	if err != nil {
		log.Print("config.GetRelayRulesFromCMD err:", err.Error())
		return
	}

	_, e := rule.AddRuleToGlobalRuleList(false, (*relayRules)...)
	if e != nil {
		log.Printf("%s\n", e)
	}

}

func LoadRuleFromConfigFile(pc *config.ProgramConfigure) {
	if pc == nil {
		return
	}
	for i := range pc.RelayRuleList {
		relayRule, err := rule.CreateRuleByConfigureAndOptions(
			pc.RelayRuleList[i].Name,
			pc.RelayRuleList[i].Configurestr,
			pc.RelayRuleList[i].Options)
		if err != nil {
			continue
		}
		relayRule.From = "configureFile" //规则来源
		relayRule.IsEnable = pc.RelayRuleList[i].Enable

		_, e := rule.AddRuleToGlobalRuleList(false, *relayRule)
		if e != nil {
			log.Printf("%s\n", e)
		}
	}
}
