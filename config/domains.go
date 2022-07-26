package config

import (
	"log"
	"strings"
	"sync"
	"time"
)

const (
	// UpdatedNothing 未改变
	UpdatedNothing updateStatusType = "域名IP和公网IP一致"
	// UpdatedFailed 更新失败
	UpdatedFailed = "失败"
	// UpdatedSuccess 更新成功
	UpdatedSuccess = "成功"
	// UpdateStop 暂停
	UpdateStop = "暂停"
	// UpdateWaiting
	UpdateWaiting = "等待更新"
)

// 固定的主域名
var staticMainDomains = []string{"com.cn", "org.cn", "net.cn", "ac.cn", "eu.org"}

// 获取ip失败的次数

// Domains Ipv4/Ipv6 DomainsState
type DomainsState struct {
	IpAddr              string
	Domains             []*Domain
	WebhookCallTime     string    `json:"WebhookCallTime"`     //最后触发时间
	WebhookCallResult   bool      `json:"WebhookCallResult"`   //触发结果
	WebhookCallErrorMsg string    `json:"WebhookCallErrorMsg"` //触发错误信息
	LastSyncTime        time.Time `json:"-"`                   //记录最新一次同步操作时间

	IPAddrHistory      []any `json:"IPAddrHistory"`
	WebhookCallHistroy []any `json:"WebhookCallHistroy"`

	Mutex *sync.RWMutex `json:"-"`
}

type IPAddrHistoryItem struct {
	IPaddr     string
	RecordTime string
}

// Domain 域名实体
type Domain struct {
	DomainName string
	SubDomain  string

	UpdateStatus         updateStatusType // 更新状态
	LastUpdateStatusTime string           //最后更新状态的时间
	Message              string

	UpdateHistroy []any

	rwmutex *sync.RWMutex
}

type UpdateHistroyItem struct {
	UpdateStatus string
	UpdateTime   string
}

type WebhookCallHistroyItem struct {
	CallTime   string
	CallResult string
}

func (d *DomainsState) SetIPAddr(ipaddr string) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	if d.IpAddr == ipaddr {
		return
	}
	d.IpAddr = ipaddr

	hi := IPAddrHistoryItem{IPaddr: ipaddr, RecordTime: time.Now().Local().Format("2006-01-02 15:04:05")}
	d.IPAddrHistory = append(d.IPAddrHistory, hi)
	if len(d.IPAddrHistory) > 10 {
		d.IPAddrHistory = DeleteAnyListlice(d.IPAddrHistory, 0)
	}
}

func (d Domain) String() string {
	if d.SubDomain != "" {
		return d.SubDomain + "." + d.DomainName
	}
	return d.DomainName
}

func DeleteAnyListlice(a []any, deleteIndex int) []any {
	j := 0
	for i := range a {
		if i != deleteIndex {
			a[j] = a[i]
			j++
		}
	}
	return a[:j]
}

func (d *DomainsState) SetDomainUpdateStatus(status updateStatusType, message string) {
	for i := range d.Domains {
		d.Domains[i].SetDomainUpdateStatus(status, message)
	}
}

func (d *DomainsState) SetWebhookResult(result bool, errMsg string) {
	d.WebhookCallResult = result
	d.WebhookCallErrorMsg = errMsg
	d.WebhookCallTime = time.Now().Format("2006-01-02 15:04:05")

	cr := "成功"
	if !result {
		cr = "出错"
	}

	hi := WebhookCallHistroyItem{CallResult: cr, CallTime: time.Now().Local().Format("2006-01-02 15:04:05")}
	d.WebhookCallHistroy = append(d.WebhookCallHistroy, hi)
	if len(d.WebhookCallHistroy) > 10 {
		d.WebhookCallHistroy = DeleteAnyListlice(d.WebhookCallHistroy, 0)
	}
}

func (d *Domain) SetDomainUpdateStatus(status updateStatusType, message string) {
	d.rwmutex.Lock()
	defer d.rwmutex.Unlock()
	d.UpdateStatus = status
	if status != UpdateWaiting && status != UpdateStop {
		d.LastUpdateStatusTime = time.Now().Format("2006-01-02 15:04:05")

		// 状态更新历史记录
		hi := UpdateHistroyItem{UpdateStatus: string(status), UpdateTime: d.LastUpdateStatusTime}
		d.UpdateHistroy = append(d.UpdateHistroy, hi)
		if len(d.UpdateHistroy) > 10 {
			d.UpdateHistroy = DeleteAnyListlice(d.UpdateHistroy, 0)
		}
	}
	d.Message = message

}

// GetFullDomain 获得全部的，子域名
func (d Domain) GetFullDomain() string {
	if d.SubDomain != "" {
		return d.SubDomain + "." + d.DomainName
	}
	return "@" + "." + d.DomainName
}

// GetSubDomain 获得子域名，为空返回@
// 阿里云，dnspod需要
func (d Domain) GetSubDomain() string {
	if d.SubDomain != "" {
		return d.SubDomain
	}
	return "@"
}

func (d *DomainsState) Init(domains []string) {
	if d.Mutex == nil {
		d.Mutex = &sync.RWMutex{}
	}

	d.Domains = d.checkParseDomains(domains)

}

// checkParseDomains 校验并解析用户输入的域名
func (d *DomainsState) checkParseDomains(domainArr []string) (domains []*Domain) {
	for _, domainStr := range domainArr {
		domainStr = strings.TrimSpace(domainStr)
		if domainStr != "" {
			dp := strings.Split(domainStr, ":")
			dplen := len(dp)
			if dplen == 1 { // 自动识别域名
				domain := &Domain{}
				domain.rwmutex = d.Mutex
				sp := strings.Split(domainStr, ".")
				length := len(sp)
				if length <= 1 {
					log.Println(domainStr, "域名不正确")
					continue
				}
				// 处理域名
				domain.DomainName = sp[length-2] + "." + sp[length-1]
				// 如包含在org.cn等顶级域名下，后三个才为用户主域名
				for _, staticMainDomain := range staticMainDomains {
					if staticMainDomain == domain.DomainName {
						domain.DomainName = sp[length-3] + "." + domain.DomainName
						break
					}
				}

				domainLen := len(domainStr) - len(domain.DomainName)
				if domainLen > 0 {
					domain.SubDomain = domainStr[:domainLen-1]
				} else {
					domain.SubDomain = domainStr[:domainLen]
				}

				domains = append(domains, domain)
			} else if dplen == 2 { // 主机记录:域名 格式
				domain := &Domain{}
				domain.rwmutex = d.Mutex
				sp := strings.Split(dp[1], ".")
				length := len(sp)
				if length <= 1 {
					log.Println(domainStr, "域名不正确")
					continue
				}
				domain.DomainName = dp[1]
				domain.SubDomain = dp[0]
				domains = append(domains, domain)
			} else {
				log.Println(domainStr, "域名不正确")
			}
		}
	}
	return
}

// CheckIPChange 检测公网IP是否改变
func (domains *DomainsState) CheckIPChange(recordType, taskKey string, getAddrFunc func() string) (ipAddr string, change bool, retDomains []*Domain) {
	ipAddr = getAddrFunc()

	checkIPChange, err := DDNSTaskIPCacheCheck(taskKey, domains.IpAddr)

	if err != nil {
		log.Printf("DDNSTaskIPCacheCheck 失败:%s", err.Error())
	}

	if err != nil || checkIPChange {
		return ipAddr, true, domains.Domains
	}

	//IP没变化
	return ipAddr, false, domains.Domains
}
