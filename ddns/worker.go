package ddns

import (
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/ddnscore.go"
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
	ddnsTaskList := ddnscore.GetDDNSTaskInfoList()
	ddnscore.CleanIPUrlAddrMap()
	ddnsConf := config.GetDDNSConfigure()

	//log.Printf("批量执行DDNS任务")
	taskBeginTime := time.Now()

	//fmt.Printf("ddnsTaskList:%v\n", ddnsTaskList)

	for index := range ddnsTaskList {

		task := ddnsTaskList[index]
		if !task.Enable {
			continue
		}

		if time.Since(task.TaskState.LastWorkTime) < time.Second*15 {
			//log.Printf("[%s]太接近,忽略", task.TaskName)
			continue
		}

		//log.Printf("task[%s] enable\n", task.TaskName)

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
			syncDDNSTask(task)
		}()

		<-time.After(time.Millisecond * 600)
	}
	wg.Wait()

	taskEndTime := time.Now()

	usedTime := taskEndTime.Sub(taskBeginTime)

	nextTaskTimer := time.Second*time.Duration(ddnsConf.Intervals) - usedTime

	//debug.FreeOSMemory()
	//log.Printf("syncAllDomainsOnce 任务完成")
	DDNSService.Timer = time.NewTimer(nextTaskTimer)
}

func syncTaskDomainsOnce(params ...any) {
	serverMsg := (params[1]).(service.ServiceMsg)
	taskKey := serverMsg.Params[0].(string)
	switch serverMsg.Type {
	case "syncDDNSTask":
		{
			//log.Printf("syncTaskDomainsOnce 单DDNS任务更新：%s", taskKey)
			ddnscore.CleanIPUrlAddrMap()
			task := ddnscore.GetDDNSTaskInfoByKey(taskKey)
			syncDDNSTask(task)
		}
	default:
		return
	}

}

func syncDDNSTask(task *ddnscore.DDNSTaskInfo) {
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
	case "porkbun":
		dnsSelected = &Porkbun{}
	default:
		return
	}

	dnsSelected.Init(task)

	dnsSelected.AddUpdateDomainRecords()
	task.ExecWebhook(&task.TaskState)
	ddnscore.DDNSTaskInfoMapUpdate(task)
	task.TaskState.LastWorkTime = time.Now() //记录最近一次检测时间,防止批量检测和单个检测时间间隔过于接近

	//
}
