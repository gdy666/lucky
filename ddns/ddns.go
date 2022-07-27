package ddns

import (
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/thirdlib/gdylib/service"
)

var DDNSService *service.Service

func init() {
	DDNSService, _ = service.NewService("ddns")
	DDNSService.SetTimerFunc(syncAllDomainsOnce)
	DDNSService.SetEventFunc(syncTaskDomainsOnce)
}

// Run 定时运行
func Run(firstDelay time.Duration, delay time.Duration) {

	log.Printf("DDNS 第一次运行将等待 %d 秒后运行 (等待网络)", int(firstDelay.Seconds()))
	<-time.After(firstDelay)
	DDNSService.Start()
}

var wg sync.WaitGroup

// RunOnce RunOnce
func syncAllDomainsOnce(params ...any) {
	ddnsTaskList := config.GetDDNSTaskList()
	config.CleanIPUrlAddrMap()
	ddnsConf := config.GetDDNSConfigure()

	for index := range ddnsTaskList {

		task := ddnsTaskList[index]
		if !task.Enable {
			config.UpdateDomainsStateByTaskKey(task.TaskKey, config.UpdateStop, "")
			continue
		}
		wg.Add(1)

		go func() {
			defer func() {
				wg.Done()
				recoverErr := recover()
				if recoverErr == nil {
					return
				}
				log.Printf("syncDDNSTask[%s]panic:\n%v", task.TaskName, recoverErr)
				log.Printf("%s", debug.Stack())
			}()
			syncDDNSTask(&task)
		}()

		<-time.After(time.Second)
	}
	wg.Wait()

	//log.Printf("syncAllDomainsOnce 任务完成")
	DDNSService.Timer = time.NewTimer(time.Second * time.Duration(ddnsConf.Intervals))
}

func syncTaskDomainsOnce(params ...any) {
	serverMsg := (params[1]).(service.ServiceMsg)
	taskKey := serverMsg.Params[0].(string)
	switch serverMsg.Type {
	case "syncDDNSTask":
		{
			//log.Printf("syncTaskDomainsOnce 单DDNS任务更新：%s", taskKey)
			task := config.GetDDNSTaskByKey(taskKey)
			syncDDNSTask(task)
		}
	default:
		return
	}

}

func syncDDNSTask(task *config.DDNSTaskDetails) {
	if task == nil {
		return
	}
	var dnsSelected DNS
	switch task.DNS.Name {
	case "alidns":
		dnsSelected = &Alidns{}
	case "dnspod":
		dnsSelected = &Dnspod{}
	case "cloudflare":
		dnsSelected = &Cloudflare{}
	case "huaweicloud":
		dnsSelected = &Huaweicloud{}
	case "callback":
		dnsSelected = &Callback{}
	case "baiducloud":
		dnsSelected = &BaiduCloud{}
	default:
		return
	}

	dnsSelected.Init(&task.DDNSTask)
	dnsSelected.AddUpdateDomainRecords()

	//task.DomainsState.IpAddr = ipaddr
	task.ExecWebhook(&task.DomainsState)

	config.DDNSTaskListFlushDomainsDetails(task.TaskKey, &task.DomainsState)
}
