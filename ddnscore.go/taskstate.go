package ddnscore

import (
	"log"
	"net/url"
	"strings"
	"time"
)

// 固定的主域名
var staticMainDomains = []string{"com.cn", "org.cn", "net.cn", "ac.cn", "eu.org"}

// 获取ip失败的次数

// Domains Ipv4/Ipv6 DDNSTaskState
type DDNSTaskState struct {
	IpAddr              string
	Domains             []Domain
	WebhookCallTime     string    `json:"WebhookCallTime"`     //最后触发时间
	WebhookCallResult   bool      `json:"WebhookCallResult"`   //触发结果
	WebhookCallErrorMsg string    `json:"WebhookCallErrorMsg"` //触发错误信息
	LastSyncTime        time.Time `json:"-"`                   //记录最新一次同步操作时间
	LastWorkTime        time.Time `json:"-"`

	IPAddrHistory      []any `json:"IPAddrHistory"`
	WebhookCallHistroy []any `json:"WebhookCallHistroy"`
}

type IPAddrHistoryItem struct {
	IPaddr     string
	RecordTime string
}

type WebhookCallHistroyItem struct {
	CallTime   string
	CallResult string
}

func (d *DDNSTaskState) SetIPAddr(ipaddr string) {
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

func (d *DDNSTaskState) SetDomainUpdateStatus(status string, message string) {
	for i := range d.Domains {
		d.Domains[i].SetDomainUpdateStatus(status, message)
	}
}

func (d *DDNSTaskState) SetWebhookResult(result bool, errMsg string) {
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

func (d *DDNSTaskState) Init(domains []string) {
	d.Domains = d.checkParseDomains(domains)

}

// checkParseDomains 校验并解析用户输入的域名
func (d *DDNSTaskState) checkParseDomains(domainArr []string) (domains []Domain) {
	for _, domainStr := range domainArr {
		domainStr = strings.TrimSpace(domainStr)
		if domainStr != "" {
			domain := &Domain{}

			dp := strings.Split(domainStr, ":")
			dplen := len(dp)
			if dplen == 1 { // 自动识别域名
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

			} else if dplen == 2 { // 主机记录:域名 格式
				sp := strings.Split(dp[1], ".")
				length := len(sp)
				if length <= 1 {
					log.Println(domainStr, "域名不正确")
					continue
				}
				domain.DomainName = dp[1]
				domain.SubDomain = dp[0]
			} else {
				log.Println(domainStr, "域名不正确")
				continue
			}

			// 参数条件
			if strings.Contains(domain.DomainName, "?") {
				u, err := url.Parse("http://" + domain.DomainName)
				if err != nil {
					log.Println(domainStr, "域名解析失败")
					continue
				}
				domain.DomainName = u.Host
				domain.CustomParams = u.Query().Encode()
			}
			domains = append(domains, *domain)
		}
	}
	return
}

// Check 检测IP是否有改变
func (state *DDNSTaskState) IPChangeCheck(newAddr string) bool {
	if newAddr == "" {
		return true
	}
	// 地址改变
	if state.IpAddr != newAddr {
		//log.Printf("公网地址改变:[%s]===>[%s]", d.DomainsInfo.IpAddr, newAddr)
		//domains.IpAddr = newAddr
		return true
	}

	return false
}
