package ddns

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gdy666/lucky/config"
	"github.com/gdy666/lucky/ddnscore.go"
	"github.com/gdy666/lucky/thirdlib/gdylib/httputils"
	"github.com/gdy666/lucky/thirdlib/gdylib/stringsp"
	"github.com/miekg/dns"
	"golang.org/x/net/idna"
)

type DNSCommon struct {
	//Domains                *config.Domains
	createUpdateDomainFunc func(recordType, ipaddr string, domain *ddnscore.Domain)
	task                   *ddnscore.DDNSTaskInfo
}

func (d *DNSCommon) SetCreateUpdateDomainFunc(f func(recordType, ipaddr string, domain *ddnscore.Domain)) {
	d.createUpdateDomainFunc = f
}

func (d *DNSCommon) Init(task *ddnscore.DDNSTaskInfo) {
	d.task = task
}

// 添加或更新IPv4/IPv6记录
func (d *DNSCommon) AddUpdateDomainRecords() string {
	if d.task.TaskType == "IPv6" {

		return d.addUpdateDomainRecords("AAAA")
	}
	return d.addUpdateDomainRecords("A")
}

func (d *DNSCommon) addUpdateDomainRecords(recordType string) string {
	ipAddr, change := d.task.CheckIPChange()
	defer ddnscore.DDNSTaskInfoMapUpdateDomainInfo(d.task)

	d.task.TaskState.SetIPAddr(ipAddr)
	//及时刷新IP地址显示
	ddnscore.DDNSTaskInfoMapUpdateIPInfo(d.task)

	if ipAddr == "" {
		d.task.TaskState.SetDomainUpdateStatus(ddnscore.UpdatePause, "获取公网IP失败")

		return ipAddr
	}

	//var preFaildDomains []ddnscore.Domain

	checkDoamins := d.task.TaskState.Domains

	//log.Printf("时间间隔:%d秒", time.Since(d.task.TaskState.LastSyncTime)/time.Second)

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

			domainResolverIPaddr, _ := ResolveDomainAtServerList(recordType, domain.String(), d.task.DNS.DNSServerList)
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

func (d *DNSCommon) CreateHTTPClient() (*http.Client, error) {
	ddnsGlobalConf := config.GetDDNSConfigure()

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

// ---------------------------------------------------------------------------------------------------
func ResolveDomainAtServerList(queryType, domain string, dnsServerList []string) (string, error) {

	if len(dnsServerList) == 0 {
		if queryType == "AAAA" {
			dnsServerList = config.DefaultIPv6DNSServerList
		} else {
			dnsServerList = config.DefaultIPv4DNSServerList
		}
	}

	//some name that ought to exist, does not exist (NXDOMAIN)

	querytype, querytypeOk := dns.StringToType[strings.ToUpper(queryType)]
	if !querytypeOk {
		return "", fmt.Errorf("queryType error:%s", queryType)
	}

	if strings.HasPrefix(domain, "*.") {
		randomStr := stringsp.GetRandomString(8)
		domain = strings.Replace(domain, "*", randomStr, 1)
	}

	domain = dns.Fqdn(domain)
	domain, err := idna.ToASCII(domain)
	if err != nil {
		return "", fmt.Errorf(` idna.ToASCII(domain) error:%s`, err.Error())
	}

	m := new(dns.Msg)
	m.SetQuestion(domain, querytype)
	m.MsgHdr.RecursionDesired = true

	c := new(dns.Client)
	noExistTimes := 0
	for _, dnsServer := range dnsServerList {
		c.Net = ""
		ipaddr, err := resolveDomain(m, c, dnsServer)
		if err != nil {
			//log.Printf("[%s]===>[%s][%s] ResolveDomain error:%s", dnsServer, queryType, domain, err.Error())
			if strings.Contains(err.Error(), "some name that ought to exist, does not exist (NXDOMAIN)") {
				noExistTimes++
				if noExistTimes >= 4 {
					return "", fmt.Errorf("解析域名[%s][%s]IP失败:noExistTimes", queryType, domain)
				}
			}
			continue
		}
		return ipaddr, nil
	}

	return "", fmt.Errorf("解析域名[%s][%s]IP失败", queryType, domain)
}

func resolveDomain(msg *dns.Msg, client *dns.Client, dnsServer string) (string, error) {

Redo:
	if in, _, err := client.Exchange(msg, dnsServer); err == nil { // Second return value is RTT, not used for now
		if in.MsgHdr.Truncated {
			client.Net = "tcp"
			goto Redo
		}

		switch in.MsgHdr.Rcode {
		case dns.RcodeServerFailure:
			return "", fmt.Errorf("the name server encountered an internal failure while processing this request (SERVFAIL)")
		case dns.RcodeNameError:
			return "", fmt.Errorf("some name that ought to exist, does not exist (NXDOMAIN)")
		case dns.RcodeRefused:
			return "", fmt.Errorf("the name server refuses to perform the specified operation for policy or security reasons (REFUSED)")
		default:
			//fmt.Printf("in.Answer len:%d\n", len(in.Answer))
			for _, rr := range in.Answer {
				//fmt.Printf("rr.String :%s\n", rr.String())
				return strings.Replace(rr.String(), rr.Header().String(), "", -1), nil
			}
		}
	}
	return "", fmt.Errorf("DNS server could not be reached")
}
