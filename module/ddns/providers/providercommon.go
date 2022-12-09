package providers

import (
	"log"
	"net/http"
	"time"

	"github.com/gdy666/lucky/module/ddns/ddnscore.go"
	"github.com/gdy666/lucky/module/ddns/ddnsgo"
	"github.com/gdy666/lucky/thirdlib/gdylib/dnsutils"
	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
)

type ProviderCommon struct {
	createUpdateDomainFunc func(recordType, ipaddr string, domain *ddnscore.Domain)
	task                   *ddnscore.DDNSTaskInfo
	taskKey                string
}

func (d *ProviderCommon) SetCreateUpdateDomainFunc(f func(recordType, ipaddr string, domain *ddnscore.Domain)) {
	d.createUpdateDomainFunc = f
}

func (d *ProviderCommon) Init(task *ddnscore.DDNSTaskInfo) {
	d.task = task
	d.taskKey = task.TaskKey
}

// 添加或更新IPv4/IPv6记录
func (d *ProviderCommon) AddUpdateDomainRecords() string {
	if d.task.TaskType == "IPv6" {

		return d.addUpdateDomainRecords("AAAA")
	}
	return d.addUpdateDomainRecords("A")
}

func (d *ProviderCommon) addUpdateDomainRecords(recordType string) string {
	ipAddr, change := d.task.CheckIPChange()
	defer ddnscore.DDNSTaskInfoMapUpdateDomainInfo(d.task)

	d.task.TaskState.SetIPAddr(ipAddr)
	//及时刷新IP地址显示
	ddnscore.DDNSTaskInfoMapUpdateIPInfo(d.task)

	if ipAddr == "" {
		d.task.TaskState.SetDomainUpdateStatus(ddnscore.UpdatePause, "获取公网IP失败")

		return ipAddr
	}

	checkDoamins := d.task.TaskState.Domains

	if time.Since(d.task.TaskState.LastSyncTime) > time.Second*time.Duration(d.task.DNS.ForceInterval-1) {
		//log.Printf("DDNS任务[%s]强制更新", d.task.TaskName)
		change = true
		goto sync
	}

	//设置原先状态成功的为继续成功
	//不成功的就更新
	if !change { //公网IP没有改变
		checkDoamins = []ddnscore.Domain{}
		for i := range d.task.TaskState.Domains { //如果原先状态成功或不改变就刷新时间
			if d.task.TaskState.Domains[i].UpdateStatus == ddnscore.UpdatedNothing ||
				d.task.TaskState.Domains[i].UpdateStatus == ddnscore.UpdatedSuccess {
				d.task.TaskState.Domains[i].SetDomainUpdateStatus(ddnscore.UpdatedNothing, "")
				ddnscore.DDNSTaskInfoMapUpdateDomainInfo(d.task)
				continue
			}
			checkDoamins = append(checkDoamins, d.task.TaskState.Domains[i])
		}

		if len(checkDoamins) == 0 {
			return ipAddr
		}
	}

sync:
	if change {
		syncTime := time.Now()
		defer func() {
			//记录最近一次同步操作时间
			d.task.TaskState.LastSyncTime = syncTime
		}()
	}

	for i := range checkDoamins {

		if d.createUpdateDomainFunc == nil {
			log.Printf("ddns createUpdateDomainFunc undefine")
			break
		}

		domain := getDomainItem(checkDoamins[i].String(), &d.task.TaskState.Domains)
		if domain == nil {
			log.Printf("getDomainItem nil")
			continue
		}

		if d.task.DNS.ResolverDoaminCheck {
			//<-time.After(time.Second)

			domainResolverIPaddr, _ := dnsutils.ResolveDomainAtServerList(recordType, domain.String(), d.task.DNS.DNSServerList)
			//log.Printf("domain:%s domainResolverIPaddr:%s ,ipaddr:%s", domain.String(), domainResolverIPaddr, ipAddr)

			if domainResolverIPaddr == ipAddr {
				if domain.UpdateStatus == ddnscore.UpdatedFailed {
					domain.SetDomainUpdateStatus(ddnscore.UpdatedSuccess, "")
				} else {
					domain.SetDomainUpdateStatus(ddnscore.UpdatedNothing, "")
				}
				ddnscore.DDNSTaskInfoMapUpdateDomainInfo(d.task)
				continue
			}
		}

		//*********
		// params := domain.GetCustomParams()
		// if params.Has("recordType") {
		// 	recordType = params.Get("recordType")
		// }

		// if params.Has("recordContent") {
		// 	//ipAddr = params.Get("recordContent")
		// 	recordContent := params.Get("recordContent")
		// 	recordContent = strings.Replace(recordContent, "#{ip}", ipAddr, -1)
		// 	ipAddr = recordContent

		// 	log.Printf("recordType[%s]recordContent[%s]", recordType, recordContent)
		// }
		//*********

		d.createUpdateDomainFunc(recordType, ipAddr, domain)
		ddnscore.DDNSTaskInfoMapUpdateDomainInfo(d.task)
	}

	return ipAddr
}

func getDomainItem(fullDomain string, domains *[]ddnscore.Domain) *ddnscore.Domain {
	if domains == nil {
		return nil
	}
	for i, domain := range *domains {
		if domain.String() == fullDomain {
			return &(*domains)[i]
		}
	}
	return nil
}

//--------------------------------------------------------------------------------------------------

func (d *ProviderCommon) CreateHTTPClient() (*http.Client, error) {
	ddnsGlobalConf := ddnsgo.GetDDNSConfigure()

	return httputils.CreateHttpClient(
		d.task.DNS.GetCallAPINetwork(),
		"",
		!ddnsGlobalConf.HttpClientSecureVerify,
		d.task.DNS.HttpClientProxyType,
		d.task.DNS.HttpClientProxyAddr,
		d.task.DNS.HttpClientProxyUser,
		d.task.DNS.HttpClientProxyPassword,
		time.Duration(d.task.HttpClientTimeout)*time.Second)
}
